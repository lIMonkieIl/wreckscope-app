package app

import (
	"context"
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"wreckscope/internal/racefeed"
	"wreckscope/internal/settings"

	wailsapp "github.com/wailsapp/wails/v3/pkg/application"
)

func TestServiceStartupSeedsBrandDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	settings.PathOverride = path
	t.Cleanup(func() { settings.PathOverride = "" })

	app := New(embed.FS{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := app.ServiceStartup(ctx, wailsapp.ServiceOptions{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ServiceShutdown() })

	got, err := settings.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Theme != "wreckscope" || got.Mode != "system" {
		t.Fatalf("first-run seed %+v", got)
	}
	live := app.GetSettings()
	if live.Theme != "wreckscope" || live.Mode != "system" {
		t.Fatalf("engine defaults %+v", live)
	}
}

func TestServiceStartupKeepsCompletedAppearance(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	settings.PathOverride = path
	t.Cleanup(func() { settings.PathOverride = "" })

	if err := settings.Save(path, racefeed.Settings{
		Theme: "pine", Mode: "light", SpeedUnit: "km/h", SetupCompleted: true,
	}); err != nil {
		t.Fatal(err)
	}

	app := New(embed.FS{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := app.ServiceStartup(ctx, wailsapp.ServiceOptions{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ServiceShutdown() })

	got := app.GetSettings()
	if got.Theme != "pine" || got.Mode != "light" {
		t.Fatalf("completed setup must keep saved appearance %+v", got)
	}
}

func TestServiceStartupResetsIncompleteAppearance(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	settings.PathOverride = path
	t.Cleanup(func() { settings.PathOverride = "" })

	if err := settings.Save(path, racefeed.Settings{
		Theme: "pine", Mode: "light", SpeedUnit: "km/h", SetupCompleted: false,
	}); err != nil {
		t.Fatal(err)
	}

	app := New(embed.FS{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := app.ServiceStartup(ctx, wailsapp.ServiceOptions{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ServiceShutdown() })

	got := app.GetSettings()
	if got.Theme != "wreckscope" || got.Mode != "system" {
		t.Fatalf("incomplete first-run must reset appearance %+v", got)
	}
	disk, err := settings.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if disk.Theme != "wreckscope" || disk.Mode != "system" {
		t.Fatalf("disk appearance %+v", disk)
	}
}

func TestSaveAppearancePersistsAndMerges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	settings.PathOverride = path
	t.Cleanup(func() { settings.PathOverride = "" })

	application := New(embed.FS{})
	_, err := application.SaveSettings(racefeed.Settings{
		SpeedUnit: "km/h",
		Theme:     "crimson",
		Mode:      "dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	before := application.GetSettings()
	if before.CautionSeconds == 0 {
		t.Fatal("expected detection defaults")
	}

	saved, err := application.SaveAppearance("mint", "system")
	if err != nil {
		t.Fatal(err)
	}
	if saved.Theme != "mint" || saved.Mode != "system" {
		t.Fatalf("appearance %+v", saved)
	}
	if saved.CautionSeconds != before.CautionSeconds {
		t.Fatalf("detection changed: %v -> %v", before.CautionSeconds, saved.CautionSeconds)
	}

	loaded, err := settings.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Theme != "mint" || loaded.Mode != "system" {
		t.Fatalf("disk appearance %+v", loaded)
	}
	if loaded.CautionSeconds != before.CautionSeconds {
		t.Fatalf("disk detection %+v", loaded)
	}
}

func TestLoadSettingsOnStartup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	settings.PathOverride = path
	t.Cleanup(func() { settings.PathOverride = "" })

	if err := settings.Save(path, racefeed.Settings{
		SpeedUnit: "mph",
		Theme:     "dracula",
		Mode:      "light",
	}); err != nil {
		t.Fatal(err)
	}

	application := New(embed.FS{})
	loaded, err := settings.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = application.engine.ApplySettings(loaded, nil)
	got := application.GetSettings()
	if got.Theme != "dracula" || got.Mode != "light" || got.SpeedUnit != "mph" {
		t.Fatalf("got %+v", got)
	}
}

func TestSetupCompletedGatesWizard(t *testing.T) {
	dir := t.TempDir()
	settings.PathOverride = filepath.Join(dir, "settings.json")
	t.Cleanup(func() { settings.PathOverride = "" })

	application := New(embed.FS{})
	status := application.SetupStatus()
	if status["setupCompleted"] != false || status["needsWizard"] != true {
		t.Fatalf("fresh setup status=%v", status)
	}

	done, err := application.CompleteSetup()
	if err != nil {
		t.Fatal(err)
	}
	if done["setupCompleted"] != true || done["needsWizard"] != false {
		t.Fatalf("after complete=%v", done)
	}

	saved, err := application.SaveSettings(racefeed.Settings{SpeedUnit: "mph", Theme: "pine", Mode: "light"})
	if err != nil {
		t.Fatal(err)
	}
	if !saved.SetupCompleted {
		t.Fatal("SaveSettings wiped setupCompleted")
	}
	raw, err := os.ReadFile(settings.PathOverride)
	if err != nil {
		t.Fatal(err)
	}
	var disk racefeed.Settings
	if err := json.Unmarshal(raw, &disk); err != nil {
		t.Fatal(err)
	}
	if !disk.SetupCompleted {
		t.Fatalf("disk missing setupCompleted: %s", raw)
	}
}
