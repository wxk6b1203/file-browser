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

	"github.com/wxk6b1203/file-util-manager/config"
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
	finalizeSeq     int64
	finalizers      map[string]*pendingDirectoryFinalize
	taskFinalizers  map[string][]string
}

type uploadDirectoryPlan struct {
	localPath  string
	remotePath string
	modTime    *time.Time
}

type uploadFilePlan struct {
	localPath  string
	remotePath string
}

type downloadDirectoryPlan struct {
	localPath   string
	remotePath  string
	sourceMTime *time.Time
}

type downloadFilePlan struct {
	localPath  string
	remotePath string
}

type directoryModTimeSource struct {
	modTime  *time.Time
	explicit bool
}

type localUploadPlan struct {
	directories []uploadDirectoryPlan
	files       []uploadFilePlan
}

type localDownloadPlan struct {
	rootPath    string
	directories []downloadDirectoryPlan
	files       []downloadFilePlan
}

type crossTransferDirectoryPlan struct {
	targetRemotePath string
	sourceMTime      *time.Time
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
	finalizeID       string
}

type directoryFinalizeKind string

const (
	finalizeLocalDirectories  directoryFinalizeKind = "local"
	finalizeRemoteDirectories directoryFinalizeKind = "remote"
)

type directoryFinalizeEntry struct {
	path    string
	modTime *time.Time
}

type pendingDirectoryFinalize struct {
	kind         directoryFinalizeKind
	connectionID string
	directories  []directoryFinalizeEntry
	expected     int
	settled      int
	active       map[string]struct{}
	failed       bool
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
		finalizers:      make(map[string]*pendingDirectoryFinalize),
		taskFinalizers:  make(map[string][]string),
	}
	svc.manager.SetObserver(svc.onManagerEvent)
	return svc
}

func newDownloadRequest(remotePath, localPath string) *folder.TransferRequest {
	return &folder.TransferRequest{
		RemotePath:      remotePath,
		LocalPath:       localPath,
		PreserveModTime: true,
	}
}

func newUploadRequest(remotePath, localPath string) *folder.TransferRequest {
	return &folder.TransferRequest{
		RemotePath:      remotePath,
		LocalPath:       localPath,
		PreserveModTime: true,
	}
}

func cloneTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	value := *t
	return &value
}

func fileInfoModTime(info *folder.FileInfo) *time.Time {
	if info == nil {
		return nil
	}
	return folder.ResolveModTime(nil, info.Metadata, info.LastModified)
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
			if dir.localPath == "" {
				continue
			}
			if err := os.MkdirAll(dir.localPath, 0o755); err != nil {
				return nil, fmt.Errorf("create local download dir %q: %w", dir.localPath, err)
			}
		}

		taskIDs := make([]string, 0, len(plan.files))
		for _, item := range plan.files {
			taskID, err := s.manager.Submit(mgr, def.Driver, def.ID, folder.TransferDownload, newDownloadRequest(item.remotePath, item.localPath))
			if err != nil {
				return nil, err
			}
			taskIDs = append(taskIDs, taskID)
		}
		if err := s.registerLocalDirectoryFinalize(plan.directories, taskIDs); err != nil {
			return nil, err
		}

		return taskIDs, nil
	}

	localPath, err := s.buildTempFilePath(connectionID, cleanRemotePath)
	if err != nil {
		return nil, err
	}

	taskID, err := s.manager.Submit(mgr, def.Driver, def.ID, folder.TransferDownload, newDownloadRequest(cleanRemotePath, localPath))
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
			if dir.localPath == "" {
				continue
			}
			if err := os.MkdirAll(dir.localPath, 0o755); err != nil {
				return nil, fmt.Errorf("create local download dir %q: %w", dir.localPath, err)
			}
		}

		taskIDs := make([]string, 0, len(plan.files))
		for _, item := range plan.files {
			taskID, err := s.manager.Submit(mgr, def.Driver, def.ID, folder.TransferDownload, newDownloadRequest(item.remotePath, item.localPath))
			if err != nil {
				return nil, err
			}
			taskIDs = append(taskIDs, taskID)
		}
		if err := s.registerLocalDirectoryFinalize(plan.directories, taskIDs); err != nil {
			return nil, err
		}

		return taskIDs, nil
	}

	localPath, err := s.prepareFileTarget(targetDir, path.Base(cleanRemotePath))
	if err != nil {
		return nil, err
	}

	taskID, err := s.manager.Submit(mgr, def.Driver, def.ID, folder.TransferDownload, newDownloadRequest(cleanRemotePath, localPath))
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

	taskID, err := s.manager.Submit(mgr, def.Driver, def.ID, folder.TransferDownload, newDownloadRequest(cleanRemotePath, targetPath))
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

	taskID, err := s.manager.Submit(mgr, def.Driver, def.ID, folder.TransferDownload, newDownloadRequest(cleanRemotePath, localPath))
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
		if dir.remotePath == "" {
			continue
		}
		if err := mgr.Mkdir(ctx, dir.remotePath); err != nil {
			return nil, err
		}
	}

	taskIDs := make([]string, 0, len(plan.files))
	for _, item := range plan.files {
		taskID, err := s.manager.Submit(mgr, def.Driver, def.ID, folder.TransferUpload, newUploadRequest(item.remotePath, item.localPath))
		if err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, taskID)
	}
	if err := s.registerRemoteDirectoryFinalize(connectionID, plan.directories, taskIDs); err != nil {
		return nil, err
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
		taskID, err := s.manager.Submit(sourceMgr, sourceDef.Driver, sourceDef.ID, folder.TransferDownload, newDownloadRequest(cleanSourcePath, localPath))
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
	items, err = enrichDirectoryItems(ctx, sourceMgr, items)
	if err != nil {
		return nil, err
	}
	rootInfo, err := sourceMgr.Stat(ctx, cleanSourcePath)
	if err != nil {
		return nil, err
	}

	localRootPath, err := s.buildTempDirectoryPath(sourceConnectionID, cleanSourcePath)
	if err != nil {
		return nil, err
	}
	downloadPlan, err := buildLocalDownloadPlan(localRootPath, cleanSourcePath, rootInfo, items)
	if err != nil {
		return nil, err
	}
	crossPlan, err := buildCrossConnectionTransferPlan(cleanSourcePath, targetDir, downloadPlan, items)
	if err != nil {
		return nil, err
	}

	for _, dir := range crossPlan.directories {
		if dir.targetRemotePath == "" {
			continue
		}
		if err := targetMgr.Mkdir(ctx, dir.targetRemotePath); err != nil {
			return nil, err
		}
	}

	finalizeID, err := s.registerDeferredRemoteDirectoryFinalize(targetConnectionID, crossPlan.directories, len(crossPlan.files))
	if err != nil {
		return nil, err
	}

	taskIDs := make([]string, 0, len(crossPlan.files))
	for _, item := range crossPlan.files {
		taskID, err := s.manager.Submit(sourceMgr, sourceDef.Driver, sourceDef.ID, folder.TransferDownload, newDownloadRequest(item.sourceRemotePath, item.localPath))
		if err != nil {
			return nil, err
		}

		s.setPendingFollowUp(taskID, pendingFollowUp{
			kind:             followUpUpload,
			localPath:        item.localPath,
			targetConnection: targetConnectionID,
			targetRemotePath: item.targetRemotePath,
			finalizeID:       finalizeID,
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
		baseDir = config.DefaultTransferTempDirPath()
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
		baseDir = config.DefaultTransferTempDirPath()
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
	rootInfo, err := mgr.Stat(ctx, remoteDir)
	if err != nil {
		return nil, err
	}
	items, err := mgr.List(ctx, remoteDir, &folder.ListOptions{Recursive: true})
	if err != nil {
		return nil, err
	}
	items, err = enrichDirectoryItems(ctx, mgr, items)
	if err != nil {
		return nil, err
	}
	return buildLocalDownloadPlan(localRootPath, remoteDir, rootInfo, items)
}

func enrichDirectoryItems(ctx context.Context, mgr folder.Manager, items []*folder.FileInfo) ([]*folder.FileInfo, error) {
	if len(items) == 0 {
		return items, nil
	}

	result := make([]*folder.FileInfo, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		if !item.IsDir() {
			result = append(result, item)
			continue
		}

		stat, err := mgr.Stat(ctx, cleanRemotePath(item.Path))
		if err == nil && stat != nil {
			result = append(result, stat)
			continue
		}
		result = append(result, item)
	}
	return result, nil
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
		directories: []uploadDirectoryPlan{},
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
		rootModTime := localInfo.ModTime()
		plan.directories = append(plan.directories, uploadDirectoryPlan{
			localPath:  cleanLocalPath,
			remotePath: targetBaseRemotePath,
			modTime:    &rootModTime,
		})
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
			dirInfo, err := entry.Info()
			if err != nil {
				return err
			}
			if targetPath != "" {
				modTime := dirInfo.ModTime()
				plan.directories = append(plan.directories, uploadDirectoryPlan{
					localPath:  currentPath,
					remotePath: targetPath,
					modTime:    &modTime,
				})
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

	sort.Slice(plan.directories, func(i, j int) bool {
		return plan.directories[i].remotePath < plan.directories[j].remotePath
	})
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

	rootInfo, err := mgr.Stat(ctx, remoteDir)
	if err != nil {
		return nil, err
	}

	items, err := mgr.List(ctx, remoteDir, &folder.ListOptions{Recursive: true})
	if err != nil {
		return nil, err
	}
	items, err = enrichDirectoryItems(ctx, mgr, items)
	if err != nil {
		return nil, err
	}

	return buildLocalDownloadPlan(localRootPath, remoteDir, rootInfo, items)
}

func buildLocalDownloadPlan(localRootPath, remoteDir string, rootInfo *folder.FileInfo, items []*folder.FileInfo) (*localDownloadPlan, error) {
	cleanRootPath := strings.TrimSpace(localRootPath)
	if cleanRootPath == "" {
		return nil, fmt.Errorf("local root path is required")
	}

	cleanRemoteDir := cleanRemotePath(remoteDir)
	if cleanRemoteDir == "" {
		return nil, fmt.Errorf("remote directory is required")
	}

	plan := &localDownloadPlan{
		rootPath: cleanRootPath,
		directories: []downloadDirectoryPlan{{
			localPath:   cleanRootPath,
			remotePath:  cleanRemoteDir,
			sourceMTime: fileInfoModTime(rootInfo),
		}},
		files: []downloadFilePlan{},
	}
	directoryMtimes := map[string]directoryModTimeSource{
		cleanRemoteDir: {
			modTime:  cloneTime(fileInfoModTime(rootInfo)),
			explicit: fileInfoModTime(rootInfo) != nil,
		},
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
			mergeDirectoryModTime(directoryMtimes, cleanItemPath, fileInfoModTime(item), true)
			continue
		}

		registerVirtualDirectories(directoryMtimes, cleanRemoteDir, cleanItemPath, fileInfoModTime(item))

		plan.files = append(plan.files, downloadFilePlan{
			localPath:  localPath,
			remotePath: cleanItemPath,
		})
	}

	for remotePath, source := range directoryMtimes {
		if remotePath == cleanRemoteDir {
			plan.directories[0].sourceMTime = cloneTime(source.modTime)
			continue
		}

		relativePath, err := filepath.Rel(cleanRemoteDir, remotePath)
		if err != nil {
			return nil, err
		}
		relativePath = filepath.ToSlash(relativePath)
		if relativePath == "." {
			continue
		}

		plan.directories = append(plan.directories, downloadDirectoryPlan{
			localPath:   filepath.Join(cleanRootPath, filepath.FromSlash(relativePath)),
			remotePath:  remotePath,
			sourceMTime: cloneTime(source.modTime),
		})
	}

	sort.Slice(plan.directories, func(i, j int) bool {
		return plan.directories[i].localPath < plan.directories[j].localPath
	})
	sort.Slice(plan.files, func(i, j int) bool {
		return plan.files[i].remotePath < plan.files[j].remotePath
	})

	return plan, nil
}

func registerVirtualDirectories(directoryMtimes map[string]directoryModTimeSource, rootDir, filePath string, fallback *time.Time) {
	parent := cleanRemotePath(path.Dir(filePath))
	rootDir = cleanRemotePath(rootDir)
	for parent != "" && parent != "." {
		mergeDirectoryModTime(directoryMtimes, parent, fallback, false)
		if parent == rootDir {
			return
		}
		nextParent := cleanRemotePath(path.Dir(parent))
		if nextParent == parent {
			return
		}
		parent = nextParent
	}
}

func mergeDirectoryModTime(directoryMtimes map[string]directoryModTimeSource, remotePath string, candidate *time.Time, explicit bool) {
	cleanPath := cleanRemotePath(remotePath)
	if cleanPath == "" {
		return
	}

	current, exists := directoryMtimes[cleanPath]
	if explicit {
		if candidate != nil || !exists {
			directoryMtimes[cleanPath] = directoryModTimeSource{
				modTime:  cloneTime(candidate),
				explicit: candidate != nil,
			}
		}
		return
	}
	if exists && current.explicit {
		return
	}
	if !exists || current.modTime == nil {
		directoryMtimes[cleanPath] = directoryModTimeSource{modTime: cloneTime(candidate)}
		return
	}
	if candidate == nil {
		return
	}
	if candidate.After(*current.modTime) {
		directoryMtimes[cleanPath] = directoryModTimeSource{modTime: cloneTime(candidate)}
	}
}

type crossTransferFilePlan struct {
	localPath        string
	sourceRemotePath string
	targetRemotePath string
}

type crossTransferPlan struct {
	directories []crossTransferDirectoryPlan
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
		directories: []crossTransferDirectoryPlan{},
		files:       []crossTransferFilePlan{},
	}
	if remoteBasePath != "" {
		rootModTime := (*time.Time)(nil)
		if len(downloadPlan.directories) > 0 {
			rootModTime = cloneTime(downloadPlan.directories[0].sourceMTime)
		}
		plan.directories = append(plan.directories, crossTransferDirectoryPlan{
			targetRemotePath: remoteBasePath,
			sourceMTime:      rootModTime,
		})
	}

	itemByPath := make(map[string]*folder.FileInfo, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		itemByPath[cleanRemotePath(item.Path)] = item
	}

	for _, localDir := range downloadPlan.directories {
		if localDir.localPath == "" || localDir.localPath == downloadPlan.rootPath {
			continue
		}
		relativePath, err := filepath.Rel(downloadPlan.rootPath, localDir.localPath)
		if err != nil {
			return nil, err
		}
		relativePath = filepath.ToSlash(relativePath)
		targetPath := joinRemotePath(remoteBasePath, relativePath)
		if targetPath != "" {
			plan.directories = append(plan.directories, crossTransferDirectoryPlan{
				targetRemotePath: targetPath,
				sourceMTime:      cloneTime(localDir.sourceMTime),
			})
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

	sort.Slice(plan.directories, func(i, j int) bool {
		return plan.directories[i].targetRemotePath < plan.directories[j].targetRemotePath
	})
	sort.Slice(plan.files, func(i, j int) bool {
		return plan.files[i].sourceRemotePath < plan.files[j].sourceRemotePath
	})

	return plan, nil
}

func (s *Service) onManagerEvent(event folder.TransferEvent) {
	if event.Type == folder.TransferEventUpsert && event.Task != nil {
		s.processFollowUp(event.Task)
		s.processDirectoryFinalizers(event.Task)
	}

	s.mu.RLock()
	observer := s.observer
	s.mu.RUnlock()
	if observer != nil {
		observer(event)
	}
}

func (s *Service) registerLocalDirectoryFinalize(directories []downloadDirectoryPlan, taskIDs []string) error {
	entries := make([]directoryFinalizeEntry, 0, len(directories))
	for _, dir := range directories {
		if dir.localPath == "" || dir.sourceMTime == nil {
			continue
		}
		entries = append(entries, directoryFinalizeEntry{
			path:    dir.localPath,
			modTime: cloneTime(dir.sourceMTime),
		})
	}
	return s.registerDirectoryFinalize(finalizeLocalDirectories, "", entries, len(taskIDs), taskIDs)
}

func (s *Service) registerRemoteDirectoryFinalize(connectionID string, directories []uploadDirectoryPlan, taskIDs []string) error {
	entries := make([]directoryFinalizeEntry, 0, len(directories))
	for _, dir := range directories {
		if dir.remotePath == "" || dir.modTime == nil {
			continue
		}
		entries = append(entries, directoryFinalizeEntry{
			path:    dir.remotePath,
			modTime: cloneTime(dir.modTime),
		})
	}
	return s.registerDirectoryFinalize(finalizeRemoteDirectories, connectionID, entries, len(taskIDs), taskIDs)
}

func (s *Service) registerDeferredRemoteDirectoryFinalize(connectionID string, directories []crossTransferDirectoryPlan, expected int) (string, error) {
	entries := make([]directoryFinalizeEntry, 0, len(directories))
	for _, dir := range directories {
		if dir.targetRemotePath == "" || dir.sourceMTime == nil {
			continue
		}
		entries = append(entries, directoryFinalizeEntry{
			path:    dir.targetRemotePath,
			modTime: cloneTime(dir.sourceMTime),
		})
	}
	return s.registerDirectoryFinalizeID(finalizeRemoteDirectories, connectionID, entries, expected, nil)
}

func (s *Service) registerDirectoryFinalize(kind directoryFinalizeKind, connectionID string, directories []directoryFinalizeEntry, expected int, taskIDs []string) error {
	_, err := s.registerDirectoryFinalizeID(kind, connectionID, directories, expected, taskIDs)
	return err
}

func (s *Service) registerDirectoryFinalizeID(kind directoryFinalizeKind, connectionID string, directories []directoryFinalizeEntry, expected int, taskIDs []string) (string, error) {
	entries := normalizeDirectoryFinalizeEntries(directories)
	if len(entries) == 0 {
		return "", nil
	}
	if expected == 0 {
		return "", s.applyDirectoryFinalize(&pendingDirectoryFinalize{
			kind:         kind,
			connectionID: connectionID,
			directories:  entries,
		})
	}

	s.mu.Lock()
	s.finalizeSeq++
	finalizeID := fmt.Sprintf("dir-finalize:%d", s.finalizeSeq)
	finalizer := &pendingDirectoryFinalize{
		kind:         kind,
		connectionID: connectionID,
		directories:  entries,
		expected:     expected,
		active:       make(map[string]struct{}, len(taskIDs)),
	}
	for _, taskID := range taskIDs {
		finalizer.active[taskID] = struct{}{}
		s.taskFinalizers[taskID] = append(s.taskFinalizers[taskID], finalizeID)
	}
	s.finalizers[finalizeID] = finalizer
	s.mu.Unlock()
	s.reconcileCompletedTasks(taskIDs)
	return finalizeID, nil
}

func normalizeDirectoryFinalizeEntries(entries []directoryFinalizeEntry) []directoryFinalizeEntry {
	result := make([]directoryFinalizeEntry, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if entry.path == "" || entry.modTime == nil {
			continue
		}
		if _, ok := seen[entry.path]; ok {
			continue
		}
		seen[entry.path] = struct{}{}
		result = append(result, directoryFinalizeEntry{
			path:    entry.path,
			modTime: cloneTime(entry.modTime),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if len(result[i].path) == len(result[j].path) {
			return result[i].path > result[j].path
		}
		return len(result[i].path) > len(result[j].path)
	})
	return result
}

func (s *Service) attachDirectoryFinalizeTask(finalizeID, taskID string) {
	if strings.TrimSpace(finalizeID) == "" || strings.TrimSpace(taskID) == "" {
		return
	}
	s.mu.Lock()
	finalizer, ok := s.finalizers[finalizeID]
	if !ok {
		s.mu.Unlock()
		return
	}
	finalizer.active[taskID] = struct{}{}
	s.taskFinalizers[taskID] = append(s.taskFinalizers[taskID], finalizeID)
	s.mu.Unlock()
	s.reconcileCompletedTasks([]string{taskID})
}

func (s *Service) reconcileCompletedTasks(taskIDs []string) {
	for _, taskID := range taskIDs {
		if strings.TrimSpace(taskID) == "" {
			continue
		}
		task := s.manager.Progress(taskID)
		if task == nil {
			continue
		}
		switch task.Status {
		case folder.TransferCompleted, folder.TransferFailed, folder.TransferCancelled:
			s.processDirectoryFinalizers(task)
		}
	}
}

func (s *Service) failDirectoryFinalize(finalizeID string) error {
	if strings.TrimSpace(finalizeID) == "" {
		return nil
	}

	var finalize *pendingDirectoryFinalize
	s.mu.Lock()
	finalizer, ok := s.finalizers[finalizeID]
	if !ok {
		s.mu.Unlock()
		return nil
	}
	finalizer.failed = true
	finalizer.settled++
	if finalizer.settled >= finalizer.expected && len(finalizer.active) == 0 {
		finalize = finalizer
		delete(s.finalizers, finalizeID)
	}
	s.mu.Unlock()
	if finalize != nil && !finalize.failed {
		return s.applyDirectoryFinalize(finalize)
	}
	return nil
}

func (s *Service) processDirectoryFinalizers(task *folder.TransferTask) {
	if task == nil || task.ID == "" {
		return
	}
	switch task.Status {
	case folder.TransferCompleted, folder.TransferFailed, folder.TransferCancelled:
	default:
		return
	}

	var finalizeList []*pendingDirectoryFinalize

	s.mu.Lock()
	finalizeIDs := append([]string(nil), s.taskFinalizers[task.ID]...)
	delete(s.taskFinalizers, task.ID)
	for _, finalizeID := range finalizeIDs {
		finalizer, ok := s.finalizers[finalizeID]
		if !ok {
			continue
		}
		delete(finalizer.active, task.ID)
		finalizer.settled++
		if task.Status != folder.TransferCompleted {
			finalizer.failed = true
		}
		if finalizer.settled >= finalizer.expected && len(finalizer.active) == 0 {
			finalizeList = append(finalizeList, finalizer)
			delete(s.finalizers, finalizeID)
		}
	}
	s.mu.Unlock()

	for _, finalize := range finalizeList {
		if finalize.failed {
			continue
		}
		if err := s.applyDirectoryFinalize(finalize); err != nil {
			s.emitErrorEvent(task.ID, err)
		}
	}
}

func (s *Service) applyDirectoryFinalize(finalizer *pendingDirectoryFinalize) error {
	if finalizer == nil || len(finalizer.directories) == 0 {
		return nil
	}

	switch finalizer.kind {
	case finalizeLocalDirectories:
		for _, entry := range finalizer.directories {
			if err := folder.ApplyLocalModTime(entry.path, entry.modTime); err != nil {
				return fmt.Errorf("restore local directory mod time for %q: %w", entry.path, err)
			}
		}
		return nil
	case finalizeRemoteDirectories:
		mgr, _, err := s.connections.Manager(context.Background(), finalizer.connectionID)
		if err != nil {
			return fmt.Errorf("prepare directory mod time restore for connection %q: %w", finalizer.connectionID, err)
		}
		setter, ok := mgr.(folder.DirectoryModTimeSetter)
		if !ok {
			return nil
		}
		for _, entry := range finalizer.directories {
			if entry.modTime == nil {
				continue
			}
			if err := setter.SetDirectoryModTime(context.Background(), entry.path, *entry.modTime); err != nil {
				return err
			}
		}
		return nil
	default:
		return nil
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
				if finalizeErr := s.failDirectoryFinalize(followUp.finalizeID); finalizeErr != nil {
					s.emitErrorEvent(task.ID, finalizeErr)
				}
				s.deletePendingFollowUp(task.ID)
				return
			}
			uploadTaskID, err := s.manager.Submit(targetMgr, targetDef.Driver, targetDef.ID, folder.TransferUpload, newUploadRequest(followUp.targetRemotePath, followUp.localPath))
			if err != nil {
				s.emitErrorEvent(task.ID, fmt.Errorf("create follow-up upload task for %q: %w", followUp.targetRemotePath, err))
				if finalizeErr := s.failDirectoryFinalize(followUp.finalizeID); finalizeErr != nil {
					s.emitErrorEvent(task.ID, finalizeErr)
				}
				s.deletePendingFollowUp(task.ID)
				return
			}
			s.attachDirectoryFinalizeTask(followUp.finalizeID, uploadTaskID)
		case followUpOpen:
			if err := s.opener(followUp.localPath); err != nil {
				s.emitErrorEvent(task.ID, fmt.Errorf("open downloaded file %q: %w", followUp.localPath, err))
				s.deletePendingFollowUp(task.ID)
				return
			}
		}
		s.deletePendingFollowUp(task.ID)
	case folder.TransferFailed, folder.TransferCancelled:
		if finalizeErr := s.failDirectoryFinalize(followUp.finalizeID); finalizeErr != nil {
			s.emitErrorEvent(task.ID, finalizeErr)
		}
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
