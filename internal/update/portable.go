package update

import (
	"os"
	"path/filepath"

	"wreckscope/internal/settings"
)

// IsPortable reports whether this process should use portable update apply
// (in-place exe swap). True when settings.json beside the exe has portable: true.
func IsPortable() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	beside := filepath.Join(filepath.Dir(exe), "settings.json")
	return settings.PortableFile(beside)
}
