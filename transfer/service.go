package transfer

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/wxk6b1203/file-util-manager/connection"
	"github.com/wxk6b1203/file-util-manager/folder"
)

const EventName = "transfer:event"

type Service struct {
	connections *connection.Service
	manager     *folder.TransferManager
	tempDir     string
	downloadDir string
	overwrite   string
	opener      func(string) error

	mu              sync.RWMutex
	observer        folder.TransferObserver
	pendingFollowUp map[string]pendingFollowUp
}

type uploadFilePlan struct {
	localPath  string
	remotePath string
}

type downloadFilePlan struct {
	localPath  string
	remotePath string
}

type localUploadPlan struct {
	directories []string
	files       []uploadFilePlan
}

type localDownloadPlan struct {
	rootPath    string
	directories []string
	files       []downloadFilePlan
}

type pendingFollowUpKind string

const (
	followUpUpload pendingFollowUpKind = "upload"
	followUpOpen   pendingFollowUpKind = "open"
)

type pendingFollowUp struct {
	kind             pendingFollowUpKind
	localPath        string
	targetConnection string
	targetRemotePath string
}

func NewService(connections *connection.Service, tempDir string, downloadDir string, overwrite string) *Service {
	svc := &Service{
		connections:     connections,
		manager:         folder.NewTransferManager(),
		tempDir:         strings.TrimSpace(tempDir),
		downloadDir:     strings.TrimSpace(downloadDir),
		overwrite:       normalizeOverwriteStrategy(overwrite),
		opener:          openLocalPath,
		pendingFollowUp: make(map[string]pendingFollowUp),
	}
	svc.manager.SetObserver(svc.onManagerEvent)
	return svc
}

func (s *Service) SetObserver(observer folder.TransferObserver) {
	s.mu.Lock()
	s.observer = observer
	s.mu.Unlock()
}

func (s *Service) SetTempDir(tempDir string) {
	s.mu.Lock()
	s.tempDir = strings.TrimSpace(tempDir)
	s.mu.Unlock()
}

func (s *Service) SetDownloadDir(downloadDir string) {
	s.mu.Lock()
	s.downloadDir = strings.TrimSpace(downloadDir)
	s.mu.Unlock()
}

func (s *Service) SetOverwriteStrategy(strategy string) {
	s.mu.Lock()
	s.overwrite = normalizeOverwriteStrategy(strategy)
	s.mu.Unlock()
}

func (s *Service) DownloadToTemp(ctx context.Context, connectionID, remotePath string) ([]string, error) {
	mgr, def, err := s.connections.Manager(ctx, connectionID)
	if err != nil {
		return nil, err
	}

	cleanRemotePath := cleanRemotePath(remotePath)
	if cleanRemotePath == "" {
		return nil, fmt.Errorf("remote path is required")
	}

	info, err := mgr.Stat(ctx, cleanRemotePath)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, fmt.Errorf("remote path %q: %w", cleanRemotePath, folder.ErrNotFound)
	}
	if info.IsDir() {
		plan, err := s.buildLocalDownloadPlan(ctx, mgr, connectionID, cleanRemotePath)
		if err != nil {
			return nil, err
		}

		for _, dir := range plan.directories {
			if dir == "" {
				continue
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("create local download dir %q: %w", dir, err)
			}
		}

		taskIDs := make([]string, 0, len(plan.files))
		for _, item := range plan.files {
			taskID, err := s.manager.Submit(mgr, def.Driver, def.ID, folder.TransferDownload, &folder.TransferRequest{
				RemotePath: item.remotePath,
				LocalPath:  item.localPath,
			})
			if err != nil {
				return nil, err
			}
			taskIDs = append(taskIDs, taskID)
		}

		return taskIDs, nil
	}

	localPath, err := s.buildTempFilePath(connectionID, cleanRemotePath)
	if err != nil {
		return nil, err
	}

	taskID, err := s.manager.Submit(mgr, def.Driver, def.ID, folder.TransferDownload, &folder.TransferRequest{
		RemotePath: cleanRemotePath,
		LocalPath:  localPath,
	})
	if err != nil {
		return nil, err
	}

	return []string{taskID}, nil
}

func (s *Service) Download(ctx context.Context, connectionID, remotePath string) ([]string, error) {
	downloadDir, ok := s.currentDownloadDir()
	if !ok {
		return s.DownloadToTemp(ctx, connectionID, remotePath)
	}
	return s.DownloadToDirectory(ctx, connectionID, remotePath, downloadDir)
}

func (s *Service) DownloadToDirectory(ctx context.Context, connectionID, remotePath, localDir string) ([]string, error) {
	mgr, def, err := s.connections.Manager(ctx, connectionID)
	if err != nil {
		return nil, err
	}

	targetDir, err := ensureLocalDirectory(localDir)
	if err != nil {
		return nil, err
	}

	cleanRemotePath := cleanRemotePath(remotePath)
	if cleanRemotePath == "" {
		return nil, fmt.Errorf("remote path is required")
	}

	info, err := mgr.Stat(ctx, cleanRemotePath)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, fmt.Errorf("remote path %q: %w", cleanRemotePath, folder.ErrNotFound)
	}

	if info.IsDir() {
		localRootPath, err := s.prepareDirectoryTarget(targetDir, cleanRemotePath)
		if err != nil {
			return nil, err
		}

		plan, err := s.buildDownloadPlanAtPath(ctx, mgr, cleanRemotePath, localRootPath)
		if err != nil {
			return nil, err
		}
		for _, dir := range plan.directories {
			if dir == "" {
				continue
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("create local download dir %q: %w", dir, err)
			}
		}

		taskIDs := make([]string, 0, len(plan.files))
		for _, item := range plan.files {
			taskID, err := s.manager.Submit(mgr, def.Driver, def.ID, folder.TransferDownload, &folder.TransferRequest{
				RemotePath: item.remotePath,
				LocalPath:  item.localPath,
			})
			if err != nil {
				return nil, err
			}
			taskIDs = append(taskIDs, taskID)
		}

		return taskIDs, nil
	}

	localPath, err := s.prepareFileTarget(targetDir, path.Base(cleanRemotePath))
	if err != nil {
		return nil, err
	}

	taskID, err := s.manager.Submit(mgr, def.Driver, def.ID, folder.TransferDownload, &folder.TransferRequest{
		RemotePath: cleanRemotePath,
		LocalPath:  localPath,
	})
	if err != nil {
		return nil, err
	}

	return []string{taskID}, nil
}

func (s *Service) DownloadFileToPath(ctx context.Context, connectionID, remotePath, localPath string) ([]string, error) {
	mgr, def, err := s.connections.Manager(ctx, connectionID)
	if err != nil {
		return nil, err
	}

	cleanRemotePath := cleanRemotePath(remotePath)
	if cleanRemotePath == "" {
		return nil, fmt.Errorf("remote path is required")
	}

	info, err := mgr.Stat(ctx, cleanRemotePath)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, fmt.Errorf("remote path %q: %w", cleanRemotePath, folder.ErrNotFound)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("remote path %q: save as only supports files", cleanRemotePath)
	}

	targetPath := strings.TrimSpace(localPath)
	if targetPath == "" {
		return []string{}, nil
	}
	if err := ensureParentDirectory(targetPath); err != nil {
		return nil, err
	}

	taskID, err := s.manager.Submit(mgr, def.Driver, def.ID, folder.TransferDownload, &folder.TransferRequest{
		RemotePath: cleanRemotePath,
		LocalPath:  targetPath,
	})
	if err != nil {
		return nil, err
	}

	return []string{taskID}, nil
}

func (s *Service) OpenFile(ctx context.Context, connectionID, remotePath string) ([]string, error) {
	mgr, def, err := s.connections.Manager(ctx, connectionID)
	if err != nil {
		return nil, err
	}

	cleanRemotePath := cleanRemotePath(remotePath)
	if cleanRemotePath == "" {
		return nil, fmt.Errorf("remote path is required")
	}

	info, err := mgr.Stat(ctx, cleanRemotePath)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, fmt.Errorf("remote path %q: %w", cleanRemotePath, folder.ErrNotFound)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("remote path %q: only files can be opened", cleanRemotePath)
	}

	localPath := ""
	if downloadDir, ok := s.currentDownloadDir(); ok {
		localPath, err = s.prepareFileTarget(downloadDir, path.Base(cleanRemotePath))
		if err != nil {
			return nil, err
		}
	} else {
		localPath, err = s.buildTempFilePath(connectionID, cleanRemotePath)
		if err != nil {
			return nil, err
		}
	}

	taskID, err := s.manager.Submit(mgr, def.Driver, def.ID, folder.TransferDownload, &folder.TransferRequest{
		RemotePath: cleanRemotePath,
		LocalPath:  localPath,
	})
	if err != nil {
		return nil, err
	}

	s.setPendingFollowUp(taskID, pendingFollowUp{
		kind:      followUpOpen,
		localPath: localPath,
	})

	return []string{taskID}, nil
}

func (s *Service) OpenLocalPath(localPath string) error {
	return openLocalPath(localPath)
}

func (s *Service) RevealLocalPath(localPath string) error {
	return revealLocalPath(localPath)
}

func (s *Service) UploadLocalPath(ctx context.Context, connectionID, remoteDir, localPath string) ([]string, error) {
	mgr, def, err := s.connections.Manager(ctx, connectionID)
	if err != nil {
		return nil, err
	}

	plan, err := buildLocalUploadPlan(remoteDir, localPath)
	if err != nil {
		return nil, err
	}

	for _, dir := range plan.directories {
		if dir == "" {
			continue
		}
		if err := mgr.Mkdir(ctx, dir); err != nil {
			return nil, err
		}
	}

	taskIDs := make([]string, 0, len(plan.files))
	for _, item := range plan.files {
		taskID, err := s.manager.Submit(mgr, def.Driver, def.ID, folder.TransferUpload, &folder.TransferRequest{
			RemotePath: item.remotePath,
			LocalPath:  item.localPath,
		})
		if err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, nil
}

func (s *Service) TransferEntry(ctx context.Context, sourceConnectionID, sourcePath, targetConnectionID, targetDir string) ([]string, error) {
	sourceMgr, sourceDef, err := s.connections.Manager(ctx, sourceConnectionID)
	if err != nil {
		return nil, err
	}
	if _, _, err := s.connections.Manager(ctx, targetConnectionID); err != nil {
		return nil, err
	}

	cleanSourcePath := cleanRemotePath(sourcePath)
	if cleanSourcePath == "" {
		return nil, fmt.Errorf("source path is required")
	}

	info, err := sourceMgr.Stat(ctx, cleanSourcePath)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, fmt.Errorf("source path %q: %w", cleanSourcePath, folder.ErrNotFound)
	}

	if !info.IsDir() {
		localPath, err := s.buildTempFilePath(sourceConnectionID, cleanSourcePath)
		if err != nil {
			return nil, err
		}

		targetRemotePath := joinRemotePath(targetDir, path.Base(cleanSourcePath))
		taskID, err := s.manager.Submit(sourceMgr, sourceDef.Driver, sourceDef.ID, folder.TransferDownload, &folder.TransferRequest{
			RemotePath: cleanSourcePath,
			LocalPath:  localPath,
		})
		if err != nil {
			return nil, err
		}

		s.setPendingFollowUp(taskID, pendingFollowUp{
			kind:             followUpUpload,
			localPath:        localPath,
			targetConnection: targetConnectionID,
			targetRemotePath: targetRemotePath,
		})
		return []string{taskID}, nil
	}

	targetMgr, _, err := s.connections.Manager(ctx, targetConnectionID)
	if err != nil {
		return nil, err
	}

	items, err := sourceMgr.List(ctx, cleanSourcePath, &folder.ListOptions{Recursive: true})
	if err != nil {
		return nil, err
	}

	localRootPath, err := s.buildTempDirectoryPath(sourceConnectionID, cleanSourcePath)
	if err != nil {
		return nil, err
	}
	downloadPlan, err := buildLocalDownloadPlan(localRootPath, cleanSourcePath, items)
	if err != nil {
		return nil, err
	}
	crossPlan, err := buildCrossConnectionTransferPlan(cleanSourcePath, targetDir, downloadPlan, items)
	if err != nil {
		return nil, err
	}

	for _, dir := range crossPlan.directories {
		if dir == "" {
			continue
		}
		if err := targetMgr.Mkdir(ctx, dir); err != nil {
			return nil, err
		}
	}

	taskIDs := make([]string, 0, len(crossPlan.files))
	for _, item := range crossPlan.files {
		taskID, err := s.manager.Submit(sourceMgr, sourceDef.Driver, sourceDef.ID, folder.TransferDownload, &folder.TransferRequest{
			RemotePath: item.sourceRemotePath,
			LocalPath:  item.localPath,
		})
		if err != nil {
			return nil, err
		}

		s.setPendingFollowUp(taskID, pendingFollowUp{
			kind:             followUpUpload,
			localPath:        item.localPath,
			targetConnection: targetConnectionID,
			targetRemotePath: item.targetRemotePath,
		})
		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, nil
}

func (s *Service) ListTasks() []*folder.TransferTask {
	return s.manager.List()
}

func (s *Service) CancelTask(taskID string) error {
	return s.manager.Cancel(strings.TrimSpace(taskID))
}

func (s *Service) RemoveTask(taskID string) error {
	return s.manager.Remove(strings.TrimSpace(taskID))
}

func (s *Service) RemoveFinishedTasks() {
	s.manager.RemoveAll()
}

func (s *Service) buildTempFilePath(connectionID, remotePath string) (string, error) {
	baseDir := s.currentTempDir()
	if baseDir == "" {
		baseDir = filepath.Join(".", "tmp", "transfers")
	}

	fileName := path.Base(remotePath)
	fileExt := path.Ext(fileName)
	fileStem := strings.TrimSuffix(fileName, fileExt)
	if fileStem == "" {
		fileStem = "download"
	}

	targetDir := filepath.Join(baseDir, "downloads", connectionID)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", fmt.Errorf("create transfer temp dir: %w", err)
	}

	timeSuffix := time.Now().Format("20060102-150405")
	targetName := fmt.Sprintf("%s-%s%s", fileStem, timeSuffix, fileExt)
	return filepath.Join(targetDir, targetName), nil
}

func (s *Service) buildTempDirectoryPath(connectionID, remotePath string) (string, error) {
	baseDir := s.currentTempDir()
	if baseDir == "" {
		baseDir = filepath.Join(".", "tmp", "transfers")
	}

	dirName := path.Base(remotePath)
	if dirName == "." || dirName == "/" || dirName == "" {
		dirName = "download"
	}

	targetDir := filepath.Join(baseDir, "downloads", connectionID)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", fmt.Errorf("create transfer temp dir: %w", err)
	}

	timeSuffix := time.Now().Format("20060102-150405")
	return filepath.Join(targetDir, fmt.Sprintf("%s-%s", dirName, timeSuffix)), nil
}

func (s *Service) buildDownloadPlanAtPath(ctx context.Context, mgr folder.Manager, remoteDir, localRootPath string) (*localDownloadPlan, error) {
	items, err := mgr.List(ctx, remoteDir, &folder.ListOptions{Recursive: true})
	if err != nil {
		return nil, err
	}
	return buildLocalDownloadPlan(localRootPath, remoteDir, items)
}

func cleanRemotePath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "." || trimmed == "/" {
		return ""
	}

	cleaned := path.Clean(strings.ReplaceAll(trimmed, "\\", "/"))
	if cleaned == "." || cleaned == "/" {
		return ""
	}
	return strings.TrimPrefix(cleaned, "/")
}

func joinRemotePath(parentDir, name string) string {
	parent := cleanRemotePath(parentDir)
	child := strings.Trim(strings.TrimSpace(name), "/")
	if parent == "" {
		return child
	}
	if child == "" {
		return parent
	}
	return path.Join(parent, child)
}

func buildLocalUploadPlan(remoteDir, localPath string) (*localUploadPlan, error) {
	cleanLocalPath := strings.TrimSpace(localPath)
	if cleanLocalPath == "" {
		return nil, fmt.Errorf("local path is required")
	}

	localInfo, err := os.Stat(cleanLocalPath)
	if err != nil {
		return nil, fmt.Errorf("stat local path %q: %w", cleanLocalPath, err)
	}

	plan := &localUploadPlan{
		directories: []string{},
		files:       []uploadFilePlan{},
	}

	if !localInfo.IsDir() {
		plan.files = append(plan.files, uploadFilePlan{
			localPath:  cleanLocalPath,
			remotePath: joinRemotePath(remoteDir, filepath.Base(cleanLocalPath)),
		})
		return plan, nil
	}

	targetBaseRemotePath := joinRemotePath(remoteDir, filepath.Base(cleanLocalPath))
	if targetBaseRemotePath != "" {
		plan.directories = append(plan.directories, targetBaseRemotePath)
	}

	err = filepath.WalkDir(cleanLocalPath, func(currentPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if currentPath == cleanLocalPath {
			return nil
		}

		relativePath, err := filepath.Rel(cleanLocalPath, currentPath)
		if err != nil {
			return err
		}

		targetPath := joinRemotePath(targetBaseRemotePath, filepath.ToSlash(relativePath))
		if entry.IsDir() {
			if targetPath != "" {
				plan.directories = append(plan.directories, targetPath)
			}
			return nil
		}

		plan.files = append(plan.files, uploadFilePlan{
			localPath:  currentPath,
			remotePath: targetPath,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk local path %q: %w", cleanLocalPath, err)
	}

	sort.Strings(plan.directories)
	sort.Slice(plan.files, func(i, j int) bool {
		return plan.files[i].remotePath < plan.files[j].remotePath
	})

	return plan, nil
}

func (s *Service) buildLocalDownloadPlan(ctx context.Context, mgr folder.Manager, connectionID, remoteDir string) (*localDownloadPlan, error) {
	localRootPath, err := s.buildTempDirectoryPath(connectionID, remoteDir)
	if err != nil {
		return nil, err
	}

	items, err := mgr.List(ctx, remoteDir, &folder.ListOptions{Recursive: true})
	if err != nil {
		return nil, err
	}

	return buildLocalDownloadPlan(localRootPath, remoteDir, items)
}

func buildLocalDownloadPlan(localRootPath, remoteDir string, items []*folder.FileInfo) (*localDownloadPlan, error) {
	cleanRootPath := strings.TrimSpace(localRootPath)
	if cleanRootPath == "" {
		return nil, fmt.Errorf("local root path is required")
	}

	cleanRemoteDir := cleanRemotePath(remoteDir)
	if cleanRemoteDir == "" {
		return nil, fmt.Errorf("remote directory is required")
	}

	plan := &localDownloadPlan{
		rootPath:    cleanRootPath,
		directories: []string{cleanRootPath},
		files:       []downloadFilePlan{},
	}

	for _, item := range items {
		if item == nil {
			continue
		}

		cleanItemPath := cleanRemotePath(item.Path)
		if cleanItemPath == "" || cleanItemPath == cleanRemoteDir {
			continue
		}

		relativePath, err := filepath.Rel(cleanRemoteDir, cleanItemPath)
		if err != nil {
			return nil, err
		}
		relativePath = filepath.ToSlash(relativePath)
		if relativePath == "." {
			continue
		}

		localPath := filepath.Join(cleanRootPath, filepath.FromSlash(relativePath))
		if item.IsDir() {
			plan.directories = append(plan.directories, localPath)
			continue
		}

		plan.files = append(plan.files, downloadFilePlan{
			localPath:  localPath,
			remotePath: cleanItemPath,
		})
	}

	sort.Strings(plan.directories)
	sort.Slice(plan.files, func(i, j int) bool {
		return plan.files[i].remotePath < plan.files[j].remotePath
	})

	return plan, nil
}

type crossTransferFilePlan struct {
	localPath        string
	sourceRemotePath string
	targetRemotePath string
}

type crossTransferPlan struct {
	directories []string
	files       []crossTransferFilePlan
}

func buildCrossConnectionTransferPlan(sourceDir, targetDir string, downloadPlan *localDownloadPlan, items []*folder.FileInfo) (*crossTransferPlan, error) {
	if downloadPlan == nil {
		return nil, fmt.Errorf("download plan is required")
	}

	cleanSourceDir := cleanRemotePath(sourceDir)
	if cleanSourceDir == "" {
		return nil, fmt.Errorf("source directory is required")
	}

	remoteBasePath := joinRemotePath(targetDir, path.Base(cleanSourceDir))
	plan := &crossTransferPlan{
		directories: []string{},
		files:       []crossTransferFilePlan{},
	}
	if remoteBasePath != "" {
		plan.directories = append(plan.directories, remoteBasePath)
	}

	itemByPath := make(map[string]*folder.FileInfo, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		itemByPath[cleanRemotePath(item.Path)] = item
	}

	for _, localDir := range downloadPlan.directories {
		if localDir == "" || localDir == downloadPlan.rootPath {
			continue
		}
		relativePath, err := filepath.Rel(downloadPlan.rootPath, localDir)
		if err != nil {
			return nil, err
		}
		relativePath = filepath.ToSlash(relativePath)
		targetPath := joinRemotePath(remoteBasePath, relativePath)
		if targetPath != "" {
			plan.directories = append(plan.directories, targetPath)
		}
	}

	for _, file := range downloadPlan.files {
		relativePath, err := filepath.Rel(downloadPlan.rootPath, file.localPath)
		if err != nil {
			return nil, err
		}
		relativePath = filepath.ToSlash(relativePath)

		sourceRemotePath := joinRemotePath(cleanSourceDir, relativePath)
		if existing, ok := itemByPath[sourceRemotePath]; ok {
			sourceRemotePath = cleanRemotePath(existing.Path)
		}

		plan.files = append(plan.files, crossTransferFilePlan{
			localPath:        file.localPath,
			sourceRemotePath: sourceRemotePath,
			targetRemotePath: joinRemotePath(remoteBasePath, relativePath),
		})
	}

	sort.Strings(plan.directories)
	sort.Slice(plan.files, func(i, j int) bool {
		return plan.files[i].sourceRemotePath < plan.files[j].sourceRemotePath
	})

	return plan, nil
}

func (s *Service) onManagerEvent(event folder.TransferEvent) {
	if event.Type == folder.TransferEventUpsert && event.Task != nil {
		s.processFollowUp(event.Task)
	}

	s.mu.RLock()
	observer := s.observer
	s.mu.RUnlock()
	if observer != nil {
		observer(event)
	}
}

func (s *Service) processFollowUp(task *folder.TransferTask) {
	if task == nil || task.ID == "" || task.Direction != folder.TransferDownload {
		return
	}

	followUp, ok := s.getPendingFollowUp(task.ID)
	if !ok {
		return
	}

	switch task.Status {
	case folder.TransferCompleted:
		switch followUp.kind {
		case followUpUpload:
			targetMgr, targetDef, err := s.connections.Manager(context.Background(), followUp.targetConnection)
			if err != nil {
				s.emitErrorEvent(task.ID, fmt.Errorf("prepare follow-up upload to connection %q: %w", followUp.targetConnection, err))
				s.deletePendingFollowUp(task.ID)
				return
			}
			_, err = s.manager.Submit(targetMgr, targetDef.Driver, targetDef.ID, folder.TransferUpload, &folder.TransferRequest{
				RemotePath: followUp.targetRemotePath,
				LocalPath:  followUp.localPath,
			})
			if err != nil {
				s.emitErrorEvent(task.ID, fmt.Errorf("create follow-up upload task for %q: %w", followUp.targetRemotePath, err))
				s.deletePendingFollowUp(task.ID)
				return
			}
		case followUpOpen:
			if err := s.opener(followUp.localPath); err != nil {
				s.emitErrorEvent(task.ID, fmt.Errorf("open downloaded file %q: %w", followUp.localPath, err))
				s.deletePendingFollowUp(task.ID)
				return
			}
		}
		s.deletePendingFollowUp(task.ID)
	case folder.TransferFailed, folder.TransferCancelled:
		s.deletePendingFollowUp(task.ID)
	}
}

func (s *Service) setPendingFollowUp(taskID string, followUp pendingFollowUp) {
	s.mu.Lock()
	s.pendingFollowUp[taskID] = followUp
	s.mu.Unlock()
}

func (s *Service) getPendingFollowUp(taskID string) (pendingFollowUp, bool) {
	s.mu.RLock()
	followUp, ok := s.pendingFollowUp[taskID]
	s.mu.RUnlock()
	return followUp, ok
}

func (s *Service) deletePendingFollowUp(taskID string) {
	s.mu.Lock()
	delete(s.pendingFollowUp, taskID)
	s.mu.Unlock()
}

func (s *Service) currentTempDir() string {
	s.mu.RLock()
	tempDir := s.tempDir
	s.mu.RUnlock()
	return strings.TrimSpace(tempDir)
}

func (s *Service) currentDownloadDir() (string, bool) {
	s.mu.RLock()
	downloadDir := strings.TrimSpace(s.downloadDir)
	s.mu.RUnlock()
	if downloadDir == "" {
		return "", false
	}

	info, err := os.Stat(downloadDir)
	if err != nil || !info.IsDir() {
		return "", false
	}

	resolved, err := filepath.EvalSymlinks(downloadDir)
	if err != nil {
		return filepath.Clean(downloadDir), true
	}
	return filepath.Clean(resolved), true
}

func (s *Service) currentOverwriteStrategy() string {
	s.mu.RLock()
	overwrite := s.overwrite
	s.mu.RUnlock()
	return normalizeOverwriteStrategy(overwrite)
}

func (s *Service) emitErrorEvent(taskID string, err error) {
	if err == nil {
		return
	}

	s.mu.RLock()
	observer := s.observer
	s.mu.RUnlock()
	if observer == nil {
		return
	}

	observer(folder.TransferEvent{
		Type:    folder.TransferEventError,
		TaskID:  taskID,
		Message: err.Error(),
	})
}

func openLocalPath(localPath string) error {
	cleanPath, _, err := existingLocalPath(localPath)
	if err != nil {
		return err
	}

	if err := commandForOpen(runtime.GOOS, cleanPath).Start(); err != nil {
		return err
	}
	return nil
}

func revealLocalPath(localPath string) error {
	cleanPath, info, err := existingLocalPath(localPath)
	if err != nil {
		return err
	}

	if err := commandForReveal(runtime.GOOS, cleanPath, info.IsDir()).Start(); err != nil {
		return err
	}
	return nil
}

func existingLocalPath(localPath string) (string, os.FileInfo, error) {
	cleanPath := strings.TrimSpace(localPath)
	if cleanPath == "" {
		return "", nil, fmt.Errorf("local path is required")
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		return "", nil, fmt.Errorf("stat local path %q: %w", cleanPath, err)
	}

	return cleanPath, info, nil
}

func commandForOpen(goos, localPath string) *exec.Cmd {
	switch goos {
	case "darwin":
		return exec.Command("open", localPath)
	case "windows":
		return exec.Command("cmd", "/c", "start", "", localPath)
	default:
		return exec.Command("xdg-open", localPath)
	}
}

func commandForReveal(goos, localPath string, isDir bool) *exec.Cmd {
	switch goos {
	case "darwin":
		if isDir {
			return exec.Command("open", localPath)
		}
		return exec.Command("open", "-R", localPath)
	case "windows":
		if isDir {
			return exec.Command("explorer", localPath)
		}
		return exec.Command("explorer", "/select,", localPath)
	default:
		targetPath := localPath
		if !isDir {
			targetPath = filepath.Dir(localPath)
		}
		return exec.Command("xdg-open", targetPath)
	}
}

func normalizeOverwriteStrategy(strategy string) string {
	if strings.EqualFold(strings.TrimSpace(strategy), "overwrite") {
		return "overwrite"
	}
	return "rename"
}

func ensureLocalDirectory(path string) (string, error) {
	target := strings.TrimSpace(path)
	if target == "" {
		return "", fmt.Errorf("local directory is required")
	}
	info, err := os.Stat(target)
	if err != nil {
		return "", fmt.Errorf("stat local directory %q: %w", target, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("local directory %q is not a directory", target)
	}
	return target, nil
}

func ensureParentDirectory(path string) error {
	parent := filepath.Dir(strings.TrimSpace(path))
	if parent == "" || parent == "." {
		return nil
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create parent directory %q: %w", parent, err)
	}
	return nil
}

func (s *Service) prepareFileTarget(baseDir, fileName string) (string, error) {
	targetDir, err := ensureLocalDirectory(baseDir)
	if err != nil {
		return "", err
	}

	targetPath := filepath.Join(targetDir, fileName)
	if s.currentOverwriteStrategy() == "overwrite" {
		return targetPath, nil
	}
	return uniqueLocalPath(targetPath), nil
}

func (s *Service) prepareDirectoryTarget(baseDir, remotePath string) (string, error) {
	targetDir, err := ensureLocalDirectory(baseDir)
	if err != nil {
		return "", err
	}

	dirName := path.Base(cleanRemotePath(remotePath))
	if dirName == "" || dirName == "." || dirName == "/" {
		dirName = "download"
	}

	targetPath := filepath.Join(targetDir, dirName)
	if s.currentOverwriteStrategy() == "overwrite" {
		return targetPath, nil
	}
	return uniqueLocalPath(targetPath), nil
}

func uniqueLocalPath(targetPath string) string {
	cleanTarget := filepath.Clean(targetPath)
	if _, err := os.Stat(cleanTarget); os.IsNotExist(err) {
		return cleanTarget
	}

	ext := filepath.Ext(cleanTarget)
	base := strings.TrimSuffix(cleanTarget, ext)
	for index := 1; ; index++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, index, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
