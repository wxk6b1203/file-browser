package transfer

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wxk6b1203/file-util-manager/folder"
)

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

	wantDirectories := []string{
		"albums/photos",
		"albums/photos/2026",
	}
	if !reflect.DeepEqual(plan.directories, wantDirectories) {
		t.Fatalf("directories = %#v, want %#v", plan.directories, wantDirectories)
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
	plan, err := buildLocalDownloadPlan(filepath.Join("tmp", "downloads", "album"), "albums/photos", []*folder.FileInfo{
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

	wantDirectories := []string{
		filepath.Join("tmp", "downloads", "album"),
		filepath.Join("tmp", "downloads", "album", "2026"),
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
	plan, err := buildLocalDownloadPlan(filepath.Join("tmp", "downloads", "empty"), "albums/empty", nil)
	if err != nil {
		t.Fatalf("buildLocalDownloadPlan error: %v", err)
	}

	wantDirectories := []string{
		filepath.Join("tmp", "downloads", "empty"),
	}
	if !reflect.DeepEqual(plan.directories, wantDirectories) {
		t.Fatalf("directories = %#v, want %#v", plan.directories, wantDirectories)
	}
	if len(plan.files) != 0 {
		t.Fatalf("files = %#v, want empty", plan.files)
	}
}

func TestBuildCrossConnectionTransferPlan_Directory(t *testing.T) {
	downloadPlan := &localDownloadPlan{
		rootPath: filepath.Join("tmp", "downloads", "photos"),
		directories: []string{
			filepath.Join("tmp", "downloads", "photos"),
			filepath.Join("tmp", "downloads", "photos", "2026"),
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

	wantDirectories := []string{
		"target/archive/photos",
		"target/archive/photos/2026",
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
