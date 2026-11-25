package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testTelemetryConfigPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "My Games", "Wreckfest 2", "steam", "savegame", "telemetry", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestEnableTelemetryForcesListenerHostPort(t *testing.T) {
	path := testTelemetryConfigPath(t)
	dir := filepath.Dir(path)
	original := []byte(`{"udp":[{"enabled":0,"ip":"10.1.2.3","port":"1234","extra":true}],"logging":[{"format":"none"}]}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if Ready(path, 23123) {
		t.Fatal("unconfigured telemetry reported ready")
	}
	if err := EnableAt(path, 23123); err != nil {
		t.Fatal(err)
	}
	if !Ready(path, 23123) {
		t.Fatal("EnableAt should force 127.0.0.1:23123 and be ready")
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(updated)
	for _, field := range []string{`"enabled": 1`, `"ip": "127.0.0.1"`, `"port": "23123"`, `"extra": true`, `"logging"`} {
		if !strings.Contains(text, field) {
			t.Errorf("missing %s in %s", field, text)
		}
	}
	if strings.Contains(text, `"10.1.2.3"`) || strings.Contains(text, `"1234"`) {
		t.Fatal("should overwrite prior host/port to the WreckScope listener")
	}
	backup, err := os.ReadFile(filepath.Join(dir, "config_backup.wreckscope.json"))
	if err != nil {
		t.Fatal("expected config_backup.wreckscope.json")
	}
	if string(backup) != string(original) {
		t.Fatal("config_backup.wreckscope.json differs from original")
	}
	if err := os.WriteFile(path, []byte(`{"udp":[{"enabled":1,"ip":"9.9.9.9","port":"9"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := RestoreBackup(path); err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(path)
	if err != nil || string(restored) != string(original) {
		t.Fatalf("restore failed: %v %s", err, restored)
	}
	if _, err := os.Stat(filepath.Join(dir, "config_backup.wreckscope.json")); !os.IsNotExist(err) {
		t.Fatal("backup should be deleted after successful restore")
	}
}

func TestEnableTelemetryFillsMissingHostPort(t *testing.T) {
	path := testTelemetryConfigPath(t)
	original := []byte(`{"udp":[{"enabled":0}]}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := EnableAt(path, 23123); err != nil {
		t.Fatal(err)
	}
	if !Ready(path, 23123) {
		t.Fatal("defaults should be ready after fill")
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(updated)
	for _, field := range []string{`"enabled": 1`, `"ip": "127.0.0.1"`, `"port": "23123"`} {
		if !strings.Contains(text, field) {
			t.Errorf("missing %s in %s", field, text)
		}
	}
}

func TestTelemetryReadySkipsWhenEnabledOnDefaults(t *testing.T) {
	path := testTelemetryConfigPath(t)
	dir := filepath.Dir(path)
	original := []byte(`{"udp":[{"enabled":1,"ip":"127.0.0.1","port":23123}]}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if !Ready(path, 23123) {
		t.Fatal("expected ready")
	}
	if err := EnableAt(path, 23123); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "config_backup.wreckscope.json")); !os.IsNotExist(err) {
		t.Fatal("already-ready config should not write a backup")
	}
}

func TestRestoreAcceptsLegacyBackupName(t *testing.T) {
	path := testTelemetryConfigPath(t)
	dir := filepath.Dir(path)
	original := []byte(`{"udp":[{"enabled":0,"ip":"127.0.0.1","port":"23123"}]}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(dir, "config-backup.json")
	if err := os.WriteFile(legacy, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"udp":[{"enabled":1,"ip":"9.9.9.9","port":"9"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := RestoreBackup(path); err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(path)
	if err != nil || string(restored) != string(original) {
		t.Fatalf("legacy restore failed: %v %s", err, restored)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatal("legacy config-backup.json should be removed after restore")
	}
	if _, err := os.Stat(filepath.Join(dir, "config_backup.wreckscope.json")); !os.IsNotExist(err) {
		t.Fatal("primary backup should not remain after restore")
	}
}

func TestRestoreAcceptsExtensionlessLegacyBackup(t *testing.T) {
	path := testTelemetryConfigPath(t)
	dir := filepath.Dir(path)
	original := []byte(`{"udp":[{"enabled":0,"ip":"127.0.0.1","port":"23123"}]}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(dir, "config-backup")
	if err := os.WriteFile(legacy, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"udp":[{"enabled":1,"ip":"9.9.9.9","port":"9"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := RestoreBackup(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatal("legacy config-backup should be removed after restore")
	}
	if _, err := os.Stat(filepath.Join(dir, "config_backup.wreckscope.json")); !os.IsNotExist(err) {
		t.Fatal("primary backup should not remain after restore")
	}
}

func TestLooksLikeTelemetryConfigHeuristics(t *testing.T) {
	good := testTelemetryConfigPath(t)
	body := []byte(`{"udp":[{"enabled":0,"ip":"127.0.0.1","port":"23123"}]}`)
	if err := os.WriteFile(good, body, 0600); err != nil {
		t.Fatal(err)
	}
	if err := LooksLikeConfig(good); err != nil {
		t.Fatalf("expected valid: %v", err)
	}

	wrongName := filepath.Join(filepath.Dir(good), "settings.json")
	if err := os.WriteFile(wrongName, body, 0600); err != nil {
		t.Fatal(err)
	}
	if err := LooksLikeConfig(wrongName); err == nil {
		t.Fatal("wrong filename should fail")
	}

	badRoot := filepath.Join(t.TempDir(), "Documents", "config.json")
	if err := os.MkdirAll(filepath.Dir(badRoot), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(badRoot, body, 0600); err != nil {
		t.Fatal(err)
	}
	if err := LooksLikeConfig(badRoot); err == nil {
		t.Fatal("path outside Wreckfest 2 should fail")
	}

	if err := os.WriteFile(good, []byte(`{"hello":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := LooksLikeConfig(good); err == nil {
		t.Fatal("missing udp should fail")
	}
}
