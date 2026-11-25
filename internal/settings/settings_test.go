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
	if PortableFile(filepath.Join(exeDir, fileName)) {
		t.Fatal("expected non-portable")
	}
}

func TestPathPortableBesideExe(t *testing.T) {
	PathOverride = ""
	exeDir := t.TempDir()
	beside := filepath.Join(exeDir, fileName)
	if err := os.WriteFile(beside, []byte(`{"portable":true,"theme":"wreckscope"}`), 0o644); err != nil {
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
	if path != beside {
		t.Fatalf("path=%q want %q", path, beside)
	}
	if !PortableFile(beside) {
		t.Fatal("expected portable")
	}
}

func TestPathPortableWithUTF8BOM(t *testing.T) {
	PathOverride = ""
	exeDir := t.TempDir()
	beside := filepath.Join(exeDir, fileName)
	bom := []byte{0xEF, 0xBB, 0xBF}
	body := append(bom, []byte(`{"portable":true,"theme":"wreckscope","mode":"system"}`)...)
	if err := os.WriteFile(beside, body, 0o644); err != nil {
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

	if !PortableFile(beside) {
		t.Fatal("BOM should not block portable detect")
	}
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if path != beside {
		t.Fatalf("path=%q want %q", path, beside)
	}
	got, err := Load(beside)
	if err != nil {
		t.Fatal(err)
	}
	if got.Theme != "wreckscope" || got.Mode != "system" || !got.Portable {
		t.Fatalf("got %+v", got)
	}
}

func TestPathIgnoresBesideExeWithoutFlag(t *testing.T) {
	PathOverride = ""
	exeDir := t.TempDir()
	configDir := t.TempDir()
	beside := filepath.Join(exeDir, fileName)
	if err := os.WriteFile(beside, []byte(`{"theme":"pine"}`), 0o644); err != nil {
		t.Fatal(err)
	}
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
}

func TestSaveKeepsPortableFlag(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, fileName)
	seed := racefeed.Settings{Portable: true, Theme: "wreckscope", Mode: "dark"}
	if err := Save(path, seed); err != nil {
		t.Fatal(err)
	}
	next := racefeed.Settings{Theme: "pine", Mode: "light", SpeedUnit: "mph"}
	if err := Save(path, next); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Portable {
		t.Fatal("portable flag dropped")
	}
	if got.Theme != "pine" {
		t.Fatalf("theme=%q", got.Theme)
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
