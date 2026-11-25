package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"wreckscope/internal/racefeed"
	"wreckscope/internal/settings"
	"wreckscope/internal/setup"
)

// legacyAppPrefsPath: old AppData prefs (telemetry path only).
func legacyAppPrefsPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "WF2 Recovery", "app.json"), nil
}

func loadLegacyTelemetryPath() string {
	path, err := legacyAppPrefsPath()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var prefs struct {
		TelemetryConfigPath string `json:"telemetryConfigPath"`
	}
	if json.Unmarshal(data, &prefs) != nil {
		return ""
	}
	return strings.TrimSpace(prefs.TelemetryConfigPath)
}

func (a *App) savedTelemetryPath() string {
	return strings.TrimSpace(a.engine.Settings().TelemetryConfigPath)
}

func (a *App) rememberTelemetryPath(path string) error {
	path = strings.TrimSpace(path)
	current := a.engine.Settings()
	if current.TelemetryConfigPath == path {
		return nil
	}
	current.TelemetryConfigPath = path
	_, err := a.persistSettings(current)
	return err
}

func (a *App) clearSavedTelemetryPath() error {
	current := a.engine.Settings()
	if current.TelemetryConfigPath == "" {
		return nil
	}
	current.TelemetryConfigPath = ""
	_, err := a.persistSettings(current)
	return err
}

// resolveTelemetryPath prefers a saved path when the file still exists.
func (a *App) resolveTelemetryPath() (string, error) {
	if saved := a.savedTelemetryPath(); saved != "" {
		if _, err := os.Stat(saved); err == nil {
			return saved, nil
		}
	}
	return setup.ConfigPath()
}

func (a *App) migrateLegacyTelemetryPath() {
	if a.savedTelemetryPath() != "" {
		return
	}
	legacy := loadLegacyTelemetryPath()
	if legacy == "" {
		return
	}
	if _, err := os.Stat(legacy); err != nil {
		return
	}
	_ = a.rememberTelemetryPath(legacy)
}

func (a *App) persistSettings(value racefeed.Settings) (racefeed.Settings, error) {
	path, pathErr := settings.Path()
	saved, err := a.engine.ApplySettings(value, func(next racefeed.Settings) error {
		if pathErr != nil {
			return pathErr
		}
		return settings.Save(path, next)
	})
	if err != nil {
		return saved, err
	}
	a.emitSettings(saved)
	a.notify()
	return saved, nil
}
