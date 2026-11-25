package update

import (
	"os"
	"path/filepath"
	"testing"

	"wreckscope/internal/settings"
)

func TestClearStagedRemovesFilesKeepsDir(t *testing.T) {
	root := t.TempDir()
	settings.PathOverride = filepath.Join(root, "settings.json")
	t.Cleanup(func() { settings.PathOverride = "" })

	dir, err := settings.UpdatesDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"WreckScope-0.1.3-windows-amd64-setup.exe", "foo.partial"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	sub := filepath.Join(dir, "keep-subdir")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}

	if err := ClearStaged(); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || !entries[0].IsDir() || entries[0].Name() != "keep-subdir" {
		t.Fatalf("after clear: %+v", entries)
	}
}

func TestClearStagedMissingDirOK(t *testing.T) {
	root := t.TempDir()
	settings.PathOverride = filepath.Join(root, "settings.json")
	t.Cleanup(func() { settings.PathOverride = "" })

	if err := ClearStaged(); err != nil {
		t.Fatal(err)
	}
}

func TestClearDirFilesBeforeDownload(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "old-setup.exe")
	if err := os.WriteFile(old, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := clearDirFiles(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatalf("old setup still present: %v", err)
	}
}
