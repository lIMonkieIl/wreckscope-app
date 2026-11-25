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

// Test hooks.
var (
	lookupExe        = os.Executable
	lookupUserConfig = os.UserConfigDir
)

// Path returns settings.json for this install mode.
// Portable: settings.json beside the exe with "portable": true.
// Installed / default: %AppData%\WreckScope\settings.json.
func Path() (string, error) {
	if PathOverride != "" {
		return PathOverride, nil
	}
	exe, err := lookupExe()
	if err != nil {
		return "", err
	}
	beside := filepath.Join(filepath.Dir(exe), fileName)
	if PortableFile(beside) {
		return beside, nil
	}
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

// PortableFile is true when path is settings.json with portable: true.
func PortableFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var probe struct {
		Portable bool `json:"portable"`
	}
	if json.Unmarshal(stripUTF8BOM(data), &probe) != nil {
		return false
	}
	return probe.Portable
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

// Save writes atomically. Keeps portable: true once set on that file.
func Save(path string, value racefeed.Settings) error {
	if existing, err := Load(path); err == nil && existing.Portable {
		value.Portable = true
	}
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
