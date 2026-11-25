package settings

import (
	"encoding/json"
	"os"
	"path/filepath"

	"wreckscope/internal/racefeed"
)

// PathOverride is for tests; empty means use the exe directory.
var PathOverride string

// Path: settings.json beside the exe.
func Path() (string, error) {
	if PathOverride != "" {
		return PathOverride, nil
	}
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), "settings.json"), nil
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
