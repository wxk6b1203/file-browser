package transfer

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/wxk6b1203/file-util-manager/connection"
	"github.com/wxk6b1203/file-util-manager/folder"
	_ "github.com/wxk6b1203/file-util-manager/folder/local"
)

type instantUploadDriver struct {
	folder.BaseDriver
}

func (d *instantUploadDriver) Capabilities() folder.Capabilities {
	caps := folder.BaseCapabilities()
	caps.CanWrite = true
	return caps
}

func (d *instantUploadDriver) Write(_ context.Context, path string, body io.Reader, _ *folder.WriteOptions) (*folder.FileInfo, error) {
	size, err := io.Copy(io.Discard, body)
	if err != nil {
		return nil, err
	}
	return &folder.FileInfo{
		Name: filepath.Base(path),
		Path: path,
		Type: folder.EntryTypeFile,
		Size: size,
	}, nil
}

func TestBuildLocalUploadPlan_File(t *testing.T) {
	root := t.TempDir()
	localFile := filepath.Join(root, "demo.txt")
	if err := os.WriteFile(localFile, []byte("demo"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	plan, err := buildLocalUploadPlan("target", localFile)
	if err != nil {
		t.Fatalf("buildLocalUploadPlan error: %v", err)
	}

	if len(plan.directories) != 0 {
		t.Fatalf("directories = %v, want empty", plan.directories)
	}

	wantFiles := []uploadFilePlan{{
		localPath:  localFile,
		remotePath: "target/demo.txt",
	}}
	if !reflect.DeepEqual(plan.files, wantFiles) {
		t.Fatalf("files = %#v, want %#v", plan.files, wantFiles)
	}
}

func TestBuildLocalUploadPlan_Directory(t *testing.T) {
	root := t.TempDir()
	localDir := filepath.Join(root, "photos")
	nestedDir := filepath.Join(localDir, "2026")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	firstFile := filepath.Join(localDir, "cover.jpg")
	secondFile := filepath.Join(nestedDir, "trip.png")
	if err := os.WriteFile(firstFile, []byte("cover"), 0o644); err != nil {
		t.Fatalf("write first file: %v", err)
	}
	if err := os.WriteFile(secondFile, []byte("trip"), 0o644); err != nil {
		t.Fatalf("write second file: %v", err)
	}

	plan, err := buildLocalUploadPlan("albums", localDir)
	if err != nil {
		t.Fatalf("buildLocalUploadPlan error: %v", err)
	}

	wantDirectoryPaths := []string{
		"albums/photos",
		"albums/photos/2026",
	}
	gotDirectoryPaths := make([]string, 0, len(plan.directories))
	for _, dir := range plan.directories {
		gotDirectoryPaths = append(gotDirectoryPaths, dir.remotePath)
		if dir.modTime == nil {
			t.Fatalf("directory %q modTime = nil", dir.remotePath)
		}
	}
	if !reflect.DeepEqual(gotDirectoryPaths, wantDirectoryPaths) {
		t.Fatalf("directories = %#v, want %#v", gotDirectoryPaths, wantDirectoryPaths)
	}

	wantFiles := []uploadFilePlan{
		{
			localPath:  secondFile,
			remotePath: "albums/photos/2026/trip.png",
		},
		{
			localPath:  firstFile,
			remotePath: "albums/photos/cover.jpg",
		},
	}
	if !reflect.DeepEqual(plan.files, wantFiles) {
		t.Fatalf("files = %#v, want %#v", plan.files, wantFiles)
	}
}

func TestBuildLocalDownloadPlan_Directory(t *testing.T) {
	plan, err := buildLocalDownloadPlan(filepath.Join("tmp", "downloads", "album"), "albums/photos", &folder.FileInfo{
		Name: "photos",
		Path: "albums/photos",
		Type: folder.EntryTypeDirectory,
	}, []*folder.FileInfo{
		{
			Name: "2026",
			Path: "albums/photos/2026",
			Type: folder.EntryTypeDirectory,
		},
		{
			Name: "trip.png",
			Path: "albums/photos/2026/trip.png",
			Type: folder.EntryTypeFile,
		},
		{
			Name: "cover.jpg",
			Path: "albums/photos/cover.jpg",
			Type: folder.EntryTypeFile,
		},
	})
	if err != nil {
		t.Fatalf("buildLocalDownloadPlan error: %v", err)
	}

	wantDirectories := []downloadDirectoryPlan{
		{
			localPath:  filepath.Join("tmp", "downloads", "album"),
			remotePath: "albums/photos",
		},
		{
			localPath:  filepath.Join("tmp", "downloads", "album", "2026"),
			remotePath: "albums/photos/2026",
		},
	}
	if !reflect.DeepEqual(plan.directories, wantDirectories) {
		t.Fatalf("directories = %#v, want %#v", plan.directories, wantDirectories)
	}

	wantFiles := []downloadFilePlan{
		{
			localPath:  filepath.Join("tmp", "downloads", "album", "2026", "trip.png"),
			remotePath: "albums/photos/2026/trip.png",
		},
		{
			localPath:  filepath.Join("tmp", "downloads", "album", "cover.jpg"),
			remotePath: "albums/photos/cover.jpg",
		},
	}
	if !reflect.DeepEqual(plan.files, wantFiles) {
		t.Fatalf("files = %#v, want %#v", plan.files, wantFiles)
	}
}

func TestBuildLocalDownloadPlan_EmptyDirectory(t *testing.T) {
	plan, err := buildLocalDownloadPlan(filepath.Join("tmp", "downloads", "empty"), "albums/empty", &folder.FileInfo{
		Name: "empty",
		Path: "albums/empty",
		Type: folder.EntryTypeDirectory,
	}, nil)
	if err != nil {
		t.Fatalf("buildLocalDownloadPlan error: %v", err)
	}

	wantDirectories := []downloadDirectoryPlan{
		{
			localPath:  filepath.Join("tmp", "downloads", "empty"),
			remotePath: "albums/empty",
		},
	}
	if !reflect.DeepEqual(plan.directories, wantDirectories) {
		t.Fatalf("directories = %#v, want %#v", plan.directories, wantDirectories)
	}
	if len(plan.files) != 0 {
		t.Fatalf("files = %#v, want empty", plan.files)
	}
}

func TestBuildLocalDownloadPlan_IncludesVirtualDirectories(t *testing.T) {
	childModTime := time.Date(2026, time.April, 4, 18, 30, 0, 0, time.UTC)
	plan, err := buildLocalDownloadPlan(filepath.Join("tmp", "downloads", "album"), "albums/photos", &folder.FileInfo{
		Name: "photos",
		Path: "albums/photos",
		Type: folder.EntryTypeDirectory,
	}, []*folder.FileInfo{
		{
			Name:         "trip.png",
			Path:         "albums/photos/2026/trip.png",
			Type:         folder.EntryTypeFile,
			LastModified: &childModTime,
		},
	})
	if err != nil {
		t.Fatalf("buildLocalDownloadPlan error: %v", err)
	}

	if len(plan.directories) != 2 {
		t.Fatalf("directories len = %d, want 2", len(plan.directories))
	}
	if got := plan.directories[1].remotePath; got != "albums/photos/2026" {
		t.Fatalf("virtual directory path = %q", got)
	}
	if plan.directories[1].sourceMTime == nil || !plan.directories[1].sourceMTime.Equal(childModTime) {
		t.Fatalf("virtual directory modTime = %v, want %v", plan.directories[1].sourceMTime, childModTime)
	}
}

func TestBuildLocalDownloadPlan_ExplicitDirectoryModTimeWins(t *testing.T) {
	markerModTime := time.Date(2026, time.April, 4, 10, 0, 0, 0, time.UTC)
	childModTime := time.Date(2026, time.April, 4, 18, 30, 0, 0, time.UTC)

	plan, err := buildLocalDownloadPlan(filepath.Join("tmp", "downloads", "album"), "albums/photos", &folder.FileInfo{
		Name: "photos",
		Path: "albums/photos",
		Type: folder.EntryTypeDirectory,
	}, []*folder.FileInfo{
		{
			Name:         "2026",
			Path:         "albums/photos/2026",
			Type:         folder.EntryTypeDirectory,
			LastModified: &markerModTime,
		},
		{
			Name:         "trip.png",
			Path:         "albums/photos/2026/trip.png",
			Type:         folder.EntryTypeFile,
			LastModified: &childModTime,
		},
	})
	if err != nil {
		t.Fatalf("buildLocalDownloadPlan error: %v", err)
	}

	dir := findDownloadDirectoryPlan(plan, "albums/photos/2026")
	if dir == nil {
		t.Fatalf("explicit directory not found in plan: %#v", plan.directories)
	}
	if dir.sourceMTime == nil || !dir.sourceMTime.Equal(markerModTime) {
		t.Fatalf("explicit directory modTime = %v, want %v", dir.sourceMTime, markerModTime)
	}
}

func TestBuildCrossConnectionTransferPlan_Directory(t *testing.T) {
	downloadPlan := &localDownloadPlan{
		rootPath: filepath.Join("tmp", "downloads", "photos"),
		directories: []downloadDirectoryPlan{
			{
				localPath:  filepath.Join("tmp", "downloads", "photos"),
				remotePath: "source/photos",
			},
			{
				localPath:  filepath.Join("tmp", "downloads", "photos", "2026"),
				remotePath: "source/photos/2026",
			},
		},
		files: []downloadFilePlan{
			{
				localPath:  filepath.Join("tmp", "downloads", "photos", "2026", "trip.png"),
				remotePath: "source/photos/2026/trip.png",
			},
			{
				localPath:  filepath.Join("tmp", "downloads", "photos", "cover.jpg"),
				remotePath: "source/photos/cover.jpg",
			},
		},
	}

	items := []*folder.FileInfo{
		{
			Name: "2026",
			Path: "source/photos/2026",
			Type: folder.EntryTypeDirectory,
		},
		{
			Name: "trip.png",
			Path: "source/photos/2026/trip.png",
			Type: folder.EntryTypeFile,
		},
		{
			Name: "cover.jpg",
			Path: "source/photos/cover.jpg",
			Type: folder.EntryTypeFile,
		},
	}

	plan, err := buildCrossConnectionTransferPlan("source/photos", "target/archive", downloadPlan, items)
	if err != nil {
		t.Fatalf("buildCrossConnectionTransferPlan error: %v", err)
	}

	wantDirectories := []crossTransferDirectoryPlan{
		{
			targetRemotePath: "target/archive/photos",
		},
		{
			targetRemotePath: "target/archive/photos/2026",
		},
	}
	if !reflect.DeepEqual(plan.directories, wantDirectories) {
		t.Fatalf("directories = %#v, want %#v", plan.directories, wantDirectories)
	}

	wantFiles := []crossTransferFilePlan{
		{
			localPath:        filepath.Join("tmp", "downloads", "photos", "2026", "trip.png"),
			sourceRemotePath: "source/photos/2026/trip.png",
			targetRemotePath: "target/archive/photos/2026/trip.png",
		},
		{
			localPath:        filepath.Join("tmp", "downloads", "photos", "cover.jpg"),
			sourceRemotePath: "source/photos/cover.jpg",
			targetRemotePath: "target/archive/photos/cover.jpg",
		},
	}
	if !reflect.DeepEqual(plan.files, wantFiles) {
		t.Fatalf("files = %#v, want %#v", plan.files, wantFiles)
	}
}

func TestProcessFollowUp_OpensDownloadedFile(t *testing.T) {
	opened := ""
	svc := &Service{
		opener: func(path string) error {
			opened = path
			return nil
		},
		pendingFollowUp: map[string]pendingFollowUp{
			"task-open": {
				kind:      followUpOpen,
				localPath: filepath.Join("tmp", "downloads", "demo.txt"),
			},
		},
	}

	svc.processFollowUp(&folder.TransferTask{
		ID:        "task-open",
		Direction: folder.TransferDownload,
		Status:    folder.TransferCompleted,
	})

	if opened != filepath.Join("tmp", "downloads", "demo.txt") {
		t.Fatalf("opened path = %q", opened)
	}
	if _, ok := svc.pendingFollowUp["task-open"]; ok {
		t.Fatalf("expected pending follow-up to be removed")
	}
}

func TestProcessFollowUpUploadRestoresDirectoryModTime(t *testing.T) {
	ctx := context.Background()
	targetRoot := t.TempDir()
	repo := connection.NewFileRepository(filepath.Join(t.TempDir(), "connections.yaml"))
	connections := connection.NewService(repo)
	svc := NewService(connections, "", "", "")

	saved, err := connections.Save(ctx, connection.Definition{
		ID:      "target-local",
		Name:    "Target Local",
		Driver:  "Local",
		Enabled: true,
		Config: map[string]any{
			"rootPath": targetRoot,
		},
	})
	if err != nil {
		t.Fatalf("save connection: %v", err)
	}

	localFile := filepath.Join(t.TempDir(), "demo.txt")
	if err := os.WriteFile(localFile, []byte("demo"), 0o644); err != nil {
		t.Fatalf("write local file: %v", err)
	}

	wantModTime := time.Date(2026, time.April, 4, 22, 23, 24, 0, time.UTC)
	finalizeID, err := svc.registerDeferredRemoteDirectoryFinalize(saved.ID, []crossTransferDirectoryPlan{{
		targetRemotePath: "target",
		sourceMTime:      &wantModTime,
	}}, 1)
	if err != nil {
		t.Fatalf("register deferred finalizer: %v", err)
	}

	svc.setPendingFollowUp("download-task", pendingFollowUp{
		kind:             followUpUpload,
		localPath:        localFile,
		targetConnection: saved.ID,
		targetRemotePath: "target/demo.txt",
		finalizeID:       finalizeID,
	})
	svc.processFollowUp(&folder.TransferTask{
		ID:        "download-task",
		Direction: folder.TransferDownload,
		Status:    folder.TransferCompleted,
	})

	targetDir := filepath.Join(targetRoot, "target")
	if err := waitForDirectoryModTime(targetDir, wantModTime, 2*time.Second); err != nil {
		t.Fatalf("wait for directory mod time: %v", err)
	}
	if _, ok := svc.getPendingFollowUp("download-task"); ok {
		t.Fatalf("expected pending follow-up to be removed")
	}
}

func TestRegisterDirectoryFinalizeReconcilesAlreadyCompletedTask(t *testing.T) {
	dirPath := t.TempDir()
	localFile := filepath.Join(t.TempDir(), "upload.txt")
	if err := os.WriteFile(localFile, []byte("demo"), 0o644); err != nil {
		t.Fatalf("write local file: %v", err)
	}
	wantModTime := time.Date(2026, time.April, 4, 19, 20, 21, 0, time.UTC)

	svc := &Service{
		manager:        folder.NewTransferManager(),
		finalizers:     make(map[string]*pendingDirectoryFinalize),
		taskFinalizers: make(map[string][]string),
	}
	taskID, err := svc.manager.Submit(&instantUploadDriver{}, "instant", "inst", folder.TransferUpload, newUploadRequest("remote/demo.txt", localFile))
	if err != nil {
		t.Fatalf("submit upload: %v", err)
	}
	if err := waitForTransferTask(svc.manager, taskID, 2*time.Second); err != nil {
		t.Fatalf("wait for task: %v", err)
	}

	if err := svc.registerDirectoryFinalize(finalizeLocalDirectories, "", []directoryFinalizeEntry{{
		path:    dirPath,
		modTime: &wantModTime,
	}}, 1, []string{taskID}); err != nil {
		t.Fatalf("register finalizer: %v", err)
	}

	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if !info.ModTime().Equal(wantModTime) {
		t.Fatalf("dir modTime = %v, want %v", info.ModTime(), wantModTime)
	}
}

func TestAttachDirectoryFinalizeTaskReconcilesAlreadyCompletedTask(t *testing.T) {
	dirPath := t.TempDir()
	localFile := filepath.Join(t.TempDir(), "upload.txt")
	if err := os.WriteFile(localFile, []byte("demo"), 0o644); err != nil {
		t.Fatalf("write local file: %v", err)
	}
	wantModTime := time.Date(2026, time.April, 4, 21, 22, 23, 0, time.UTC)

	svc := &Service{
		manager:        folder.NewTransferManager(),
		finalizers:     make(map[string]*pendingDirectoryFinalize),
		taskFinalizers: make(map[string][]string),
	}
	finalizeID, err := svc.registerDirectoryFinalizeID(finalizeLocalDirectories, "", []directoryFinalizeEntry{{
		path:    dirPath,
		modTime: &wantModTime,
	}}, 1, nil)
	if err != nil {
		t.Fatalf("register deferred finalizer: %v", err)
	}

	taskID, err := svc.manager.Submit(&instantUploadDriver{}, "instant", "inst", folder.TransferUpload, newUploadRequest("remote/demo.txt", localFile))
	if err != nil {
		t.Fatalf("submit upload: %v", err)
	}
	if err := waitForTransferTask(svc.manager, taskID, 2*time.Second); err != nil {
		t.Fatalf("wait for task: %v", err)
	}

	svc.attachDirectoryFinalizeTask(finalizeID, taskID)

	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if !info.ModTime().Equal(wantModTime) {
		t.Fatalf("dir modTime = %v, want %v", info.ModTime(), wantModTime)
	}
}

func TestCommandForOpen(t *testing.T) {
	t.Run("darwin", func(t *testing.T) {
		cmd := commandForOpen("darwin", "/tmp/demo.txt")
		if got, want := filepath.Base(cmd.Path), "open"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if !reflect.DeepEqual(cmd.Args, []string{"open", "/tmp/demo.txt"}) {
			t.Fatalf("args = %#v", cmd.Args)
		}
	})

	t.Run("windows", func(t *testing.T) {
		cmd := commandForOpen("windows", `C:\tmp\demo.txt`)
		if got, want := filepath.Base(cmd.Path), "cmd"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if !reflect.DeepEqual(cmd.Args, []string{"cmd", "/c", "start", "", `C:\tmp\demo.txt`}) {
			t.Fatalf("args = %#v", cmd.Args)
		}
	})

	t.Run("linux", func(t *testing.T) {
		cmd := commandForOpen("linux", "/tmp/demo.txt")
		if got, want := filepath.Base(cmd.Path), "xdg-open"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if !reflect.DeepEqual(cmd.Args, []string{"xdg-open", "/tmp/demo.txt"}) {
			t.Fatalf("args = %#v", cmd.Args)
		}
	})
}

func waitForTransferTask(manager *folder.TransferManager, taskID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		task := manager.Progress(taskID)
		if task == nil {
			return nil
		}
		switch task.Status {
		case folder.TransferCompleted, folder.TransferFailed, folder.TransferCancelled:
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return context.DeadlineExceeded
}

func findDownloadDirectoryPlan(plan *localDownloadPlan, remotePath string) *downloadDirectoryPlan {
	if plan == nil {
		return nil
	}
	for i := range plan.directories {
		if plan.directories[i].remotePath == remotePath {
			return &plan.directories[i]
		}
	}
	return nil
}

func waitForDirectoryModTime(dirPath string, want time.Time, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		info, err := os.Stat(dirPath)
		if err == nil && modTimeClose(info.ModTime(), want) {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return context.DeadlineExceeded
}

func modTimeClose(got, want time.Time) bool {
	diff := got.Sub(want)
	if diff < 0 {
		diff = -diff
	}
	return diff < time.Second
}

func TestCommandForReveal(t *testing.T) {
	t.Run("darwin file", func(t *testing.T) {
		cmd := commandForReveal("darwin", "/tmp/demo.txt", false)
		if got, want := filepath.Base(cmd.Path), "open"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if !reflect.DeepEqual(cmd.Args, []string{"open", "-R", "/tmp/demo.txt"}) {
			t.Fatalf("args = %#v", cmd.Args)
		}
	})

	t.Run("darwin dir", func(t *testing.T) {
		cmd := commandForReveal("darwin", "/tmp/downloads", true)
		if !reflect.DeepEqual(cmd.Args, []string{"open", "/tmp/downloads"}) {
			t.Fatalf("args = %#v", cmd.Args)
		}
	})

	t.Run("windows file", func(t *testing.T) {
		cmd := commandForReveal("windows", `C:\tmp\demo.txt`, false)
		if got, want := filepath.Base(cmd.Path), "explorer"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if !reflect.DeepEqual(cmd.Args, []string{"explorer", "/select,", `C:\tmp\demo.txt`}) {
			t.Fatalf("args = %#v", cmd.Args)
		}
	})

	t.Run("windows dir", func(t *testing.T) {
		cmd := commandForReveal("windows", `C:\tmp\downloads`, true)
		if !reflect.DeepEqual(cmd.Args, []string{"explorer", `C:\tmp\downloads`}) {
			t.Fatalf("args = %#v", cmd.Args)
		}
	})

	t.Run("linux file", func(t *testing.T) {
		cmd := commandForReveal("linux", filepath.Join("/tmp", "downloads", "demo.txt"), false)
		if got, want := filepath.Base(cmd.Path), "xdg-open"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if !reflect.DeepEqual(cmd.Args, []string{"xdg-open", filepath.Join("/tmp", "downloads")}) {
			t.Fatalf("args = %#v", cmd.Args)
		}
	})

	t.Run("linux dir", func(t *testing.T) {
		cmd := commandForReveal("linux", filepath.Join("/tmp", "downloads"), true)
		if !reflect.DeepEqual(cmd.Args, []string{"xdg-open", filepath.Join("/tmp", "downloads")}) {
			t.Fatalf("args = %#v", cmd.Args)
		}
	})
}
