package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"wreckscope/internal/racefeed"
)

func TestPathBesideExe(t *testing.T) {
	PathOverride = ""
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "settings.json" {
		t.Fatalf("base=%s", filepath.Base(path))
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(path) != filepath.Dir(exe) {
		t.Fatalf("settings dir %q != exe dir %q", filepath.Dir(path), filepath.Dir(exe))
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
