package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"wreckscope/internal/racefeed"
)

func TestPathInstalledUsesAppData(t *testing.T) {
	PathOverride = ""
	exeDir := t.TempDir()
	configDir := t.TempDir()
	lookupExe = func() (string, error) {
		return filepath.Join(exeDir, "WreckScope.exe"), nil
	}
	lookupUserConfig = func() (string, error) { return configDir, nil }
	t.Cleanup(func() {
		lookupExe = os.Executable
		lookupUserConfig = os.UserConfigDir
	})

	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(configDir, appDataFolder, fileName)
	if path != want {
		t.Fatalf("path=%q want %q", path, want)
	}
	if Portable(exeDir) {
		t.Fatal("expected non-portable")
	}
}

func TestPathPortableBesideExe(t *testing.T) {
	PathOverride = ""
	exeDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(exeDir, portableMarker), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	lookupExe = func() (string, error) {
		return filepath.Join(exeDir, "WreckScope.exe"), nil
	}
	lookupUserConfig = func() (string, error) {
		t.Fatal("UserConfigDir should not be used in portable mode")
		return "", nil
	}
	t.Cleanup(func() {
		lookupExe = os.Executable
		lookupUserConfig = os.UserConfigDir
	})

	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(exeDir, fileName)
	if path != want {
		t.Fatalf("path=%q want %q", path, want)
	}
	if !Portable(exeDir) {
		t.Fatal("expected portable")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	PathOverride = path
	t.Cleanup(func() { PathOverride = "" })

	want := racefeed.Settings{
		DrawTrackOutline: true,
		SpeedUnit:        "mph",
		MuteAlertSound:   true,
		Theme:            "pine",
		Mode:             "light",
	}
	want.CautionSeconds = 2
	want.AlertSeconds = 4
	want.SpeedThreshold = 5
	want.HideBots = true

	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %+v want %+v", got, want)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var check map[string]any
	if err := json.Unmarshal(raw, &check); err != nil {
		t.Fatal(err)
	}
	if check["theme"] != "pine" || check["mode"] != "light" {
		t.Fatalf("json appearance: %v", check)
	}
}
