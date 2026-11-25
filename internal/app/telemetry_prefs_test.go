package app

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"wreckscope/internal/racefeed"
	"wreckscope/internal/settings"
)

func TestRememberAndClearTelemetryPathInSettings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	settings.PathOverride = path
	t.Cleanup(func() { settings.PathOverride = "" })

	app := New(embed.FS{})
	config := filepath.Join(dir, "game", "config.json")
	if err := os.MkdirAll(filepath.Dir(config), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}

	if err := app.rememberTelemetryPath(config); err != nil {
		t.Fatal(err)
	}
	if app.savedTelemetryPath() != config {
		t.Fatalf("saved=%q", app.savedTelemetryPath())
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var disk racefeed.Settings
	if err := json.Unmarshal(raw, &disk); err != nil {
		t.Fatal(err)
	}
	if disk.TelemetryConfigPath != config {
		t.Fatalf("disk path=%q", disk.TelemetryConfigPath)
	}

	if err := app.clearSavedTelemetryPath(); err != nil {
		t.Fatal(err)
	}
	if app.savedTelemetryPath() != "" {
		t.Fatalf("expected cleared, got %q", app.savedTelemetryPath())
	}
}

func TestMigrateLegacyTelemetryPath(t *testing.T) {
	dir := t.TempDir()
	settingsFile := filepath.Join(dir, "settings.json")
	settings.PathOverride = settingsFile
	t.Cleanup(func() { settings.PathOverride = "" })

	t.Setenv("AppData", dir)
	config := filepath.Join(dir, "game", "config.json")
	if err := os.MkdirAll(filepath.Dir(config), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	legacyDir := filepath.Join(dir, "WF2 Recovery")
	if err := os.MkdirAll(legacyDir, 0700); err != nil {
		t.Fatal(err)
	}
	legacy := map[string]string{"telemetryConfigPath": config}
	data, _ := json.Marshal(legacy)
	if err := os.WriteFile(filepath.Join(legacyDir, "app.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	app := New(embed.FS{})
	app.migrateLegacyTelemetryPath()
	if app.savedTelemetryPath() != config {
		t.Fatalf("migrated=%q", app.savedTelemetryPath())
	}
	raw, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(raw) || !containsPath(raw, config) {
		t.Fatalf("settings.json missing path: %s", raw)
	}
}

func containsPath(raw []byte, path string) bool {
	var settings racefeed.Settings
	if json.Unmarshal(raw, &settings) != nil {
		return false
	}
	return settings.TelemetryConfigPath == path
}

func TestSaveSettingsPreservesTelemetryPath(t *testing.T) {
	dir := t.TempDir()
	settings.PathOverride = filepath.Join(dir, "settings.json")
	t.Cleanup(func() { settings.PathOverride = "" })

	app := New(embed.FS{})
	config := filepath.Join(dir, "config.json")
	if err := app.rememberTelemetryPath(config); err != nil {
		t.Fatal(err)
	}
	saved, err := app.SaveSettings(racefeed.Settings{
		SpeedUnit: "mph",
		Theme:     "pine",
		Mode:      "light",
	})
	if err != nil {
		t.Fatal(err)
	}
	if saved.TelemetryConfigPath != config {
		t.Fatalf("path wiped: %q", saved.TelemetryConfigPath)
	}
	if saved.SpeedUnit != "mph" || saved.Theme != "pine" {
		t.Fatalf("settings %+v", saved)
	}
}
