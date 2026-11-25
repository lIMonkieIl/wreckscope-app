package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"wreckscope/internal/racefeed"
)

func TestPathUsesAppData(t *testing.T) {
	PathOverride = ""
	configDir := t.TempDir()
	lookupUserConfig = func() (string, error) { return configDir, nil }
	t.Cleanup(func() {
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

func TestWebviewDataDirUnderAppData(t *testing.T) {
	PathOverride = ""
	configDir := t.TempDir()
	lookupUserConfig = func() (string, error) { return configDir, nil }
	t.Cleanup(func() {
		lookupUserConfig = os.UserConfigDir
	})

	got, err := WebviewDataDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(configDir, appDataFolder, "webview")
	if got != want {
		t.Fatalf("webview=%q want %q", got, want)
	}
}

func TestUpdatesDirUnderAppData(t *testing.T) {
	PathOverride = ""
	configDir := t.TempDir()
	lookupUserConfig = func() (string, error) { return configDir, nil }
	t.Cleanup(func() {
		lookupUserConfig = os.UserConfigDir
	})

	got, err := UpdatesDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(configDir, appDataFolder, "updates")
	if got != want {
		t.Fatalf("updates=%q want %q", got, want)
	}
}

func TestLoadSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, fileName)
	in := racefeed.Settings{Theme: "wreckscope", Mode: "system", SpeedUnit: "km/h"}
	if err := Save(path, in); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Theme != "wreckscope" || got.Mode != "system" {
		t.Fatalf("got %+v", got)
	}
}

func TestLoadStripsUTF8BOM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, fileName)
	body := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"theme":"wreckscope","mode":"system"}`)...)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Theme != "wreckscope" || got.Mode != "system" {
		t.Fatalf("got %+v", got)
	}
}

func TestSaveJSONShape(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, fileName)
	if err := Save(path, racefeed.Settings{Theme: "pine", Mode: "light"}); err != nil {
		t.Fatal(err)
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
		t.Fatalf("json=%v", check)
	}
	if _, ok := check["portable"]; ok {
		t.Fatal("portable key must not appear")
	}
}
