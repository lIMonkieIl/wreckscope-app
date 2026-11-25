package update

import (
	"os"
	"path/filepath"

	"wreckscope/internal/settings"
)

// ClearStaged removes downloaded setup packages under %AppData%\WreckScope\updates.
// Leaves the directory itself. Missing dir is fine. Best-effort per file.
func ClearStaged() error {
	dir, err := settings.UpdatesDir()
	if err != nil {
		return err
	}
	return clearDirFiles(dir)
}

func clearDirFiles(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var first error
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if err := os.Remove(path); err != nil && first == nil {
			first = err
		}
	}
	return first
}
