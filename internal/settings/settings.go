package settings

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"wreckscope/internal/racefeed"
)

const (
	fileName      = "settings.json"
	appDataFolder = "WreckScope"
)

// PathOverride is for tests.
var PathOverride string

// Test hook.
var lookupUserConfig = os.UserConfigDir

// Path returns %AppData%\WreckScope\settings.json.
// All app data lives under AppData; the install dir holds binaries only.
func Path() (string, error) {
	if PathOverride != "" {
		return PathOverride, nil
	}
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

// WebviewDataDir is %AppData%\WreckScope\webview (stable; not named after the exe).
func WebviewDataDir() (string, error) {
	if PathOverride != "" {
		return filepath.Join(filepath.Dir(PathOverride), "webview"), nil
	}
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "webview"), nil
}

// UpdatesDir is %AppData%\WreckScope\updates (downloaded setup packages).
func UpdatesDir() (string, error) {
	if PathOverride != "" {
		return filepath.Join(filepath.Dir(PathOverride), "updates"), nil
	}
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "updates"), nil
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
	err = json.Unmarshal(stripUTF8BOM(data), &value)
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

// stripUTF8BOM drops a leading UTF-8 BOM (PowerShell Set-Content -Encoding utf8).
func stripUTF8BOM(data []byte) []byte {
	return bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
}
