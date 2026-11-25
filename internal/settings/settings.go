package settings

import (
	"encoding/json"
	"os"
	"path/filepath"

	"wreckscope/internal/racefeed"
)

const (
	fileName        = "settings.json"
	portableMarker  = "portable"
	appDataFolder   = "WreckScope"
)

// PathOverride is for tests.
var PathOverride string

// Test hooks (nil = real OS).
var (
	lookupExe        = os.Executable
	lookupUserConfig = os.UserConfigDir
)

// Path returns settings.json for this install mode.
// Portable (marker file beside the exe): next to the exe.
// Installed / default: %AppData%\WreckScope\settings.json.
func Path() (string, error) {
	if PathOverride != "" {
		return PathOverride, nil
	}
	exe, err := lookupExe()
	if err != nil {
		return "", err
	}
	exeDir := filepath.Dir(exe)
	if Portable(exeDir) {
		return filepath.Join(exeDir, fileName), nil
	}
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

// Portable is true when exeDir contains the portable marker file.
func Portable(exeDir string) bool {
	_, err := os.Stat(filepath.Join(exeDir, portableMarker))
	return err == nil
}

func dataDir() (string, error) {
	root, err := lookupUserConfig()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, appDataFolder), nil
}

func Load(path string) (racefeed.Settings, error) {
	var value racefeed.Settings
	data, err := os.ReadFile(path)
	if err != nil {
		return value, err
	}
	err = json.Unmarshal(data, &value)
	return value, err
}

// Save writes atomically.
func Save(path string, value racefeed.Settings) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, "settings-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	enc := json.NewEncoder(temp)
	enc.SetIndent("", "  ")
	err = enc.Encode(value)
	if err == nil {
		err = temp.Sync()
	}
	closeErr := temp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(temp.Name(), path)
}
