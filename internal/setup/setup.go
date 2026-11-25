package setup

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func PhoneURL(port int) string {
	// UDP connect selects the default IPv4 route without sending a packet.
	conn, err := net.Dial("udp4", "192.0.2.1:9")
	if err != nil {
		return ""
	}
	defer conn.Close()
	address := conn.LocalAddr().(*net.UDPAddr)
	if !address.IP.IsPrivate() {
		return ""
	}
	return fmt.Sprintf("http://%s:%d", address.IP, port)
}

func ConfigPath() (string, error) {
	roots := []string{filepath.Join(os.Getenv("USERPROFILE"), "Documents")}
	for _, variable := range []string{"OneDrive", "OneDriveConsumer", "OneDriveCommercial"} {
		if root := os.Getenv(variable); root != "" {
			roots = append(roots, filepath.Join(root, "Documents"))
		}
	}
	var path string
	for _, root := range roots {
		matches, err := filepath.Glob(filepath.Join(root, "My Games", "Wreckfest 2", "*", "savegame", "telemetry", "config.json"))
		if err != nil {
			return "", err
		}
		for _, match := range matches {
			if path != "" && !strings.EqualFold(path, match) {
				return "", fmt.Errorf("multiple game profiles found; set telemetry manually")
			}
			path = match
		}
	}
	if path == "" {
		return "", fmt.Errorf("no telemetry config found; launch Wreckfest 2 once, close it, then reopen WreckScope")
	}
	return path, nil
}

func Ready(path string, port int) bool {
	if LooksLikeConfig(path) != nil {
		return false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var config struct {
		UDP []struct {
			Enabled int
			IP      string
			Port    json.RawMessage
		}
	}
	if json.Unmarshal(data, &config) != nil || len(config.UDP) != 1 {
		return false
	}
	target := config.UDP[0]
	expected := strconv.Itoa(port)
	return target.Enabled == 1 && target.IP == "127.0.0.1" &&
		(string(target.Port) == expected || string(target.Port) == strconv.Quote(expected))
}

// LooksLikeConfig: name/path/structure for .../Wreckfest 2/.../savegame/telemetry/config.json.
func LooksLikeConfig(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("choose the game telemetry config.json")
	}
	if !strings.EqualFold(filepath.Base(path), "config.json") {
		return fmt.Errorf("file must be named config.json")
	}
	lower := strings.ToLower(filepath.ToSlash(path))
	if !strings.Contains(lower, "/wreckfest 2/") {
		return fmt.Errorf("path must be under My Games/Wreckfest 2")
	}
	if !strings.Contains(lower, "/savegame/telemetry/config.json") {
		return fmt.Errorf("path must be .../savegame/telemetry/config.json")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var config map[string]json.RawMessage
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("not valid JSON")
	}
	rawUDP, ok := config["udp"]
	if !ok || len(rawUDP) == 0 || string(rawUDP) == "null" {
		return fmt.Errorf("missing telemetry udp section")
	}
	var targets []map[string]json.RawMessage
	if err := json.Unmarshal(rawUDP, &targets); err != nil || len(targets) != 1 || targets[0] == nil {
		return fmt.Errorf("expected one UDP destination in telemetry config")
	}
	if _, ok := targets[0]["enabled"]; !ok {
		return fmt.Errorf("UDP destination missing enabled")
	}
	return nil
}

// EnableAt sets udp[0] to enabled on 127.0.0.1:defaultPort (WreckScope's listener).
func EnableAt(path string, defaultPort int) error {
	original, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var config map[string]json.RawMessage
	if err := json.Unmarshal(original, &config); err != nil {
		return err
	}
	var targets []map[string]json.RawMessage
	if err := json.Unmarshal(config["udp"], &targets); err != nil || len(targets) != 1 || targets[0] == nil {
		return fmt.Errorf("expected one existing UDP destination")
	}
	target := targets[0]

	ip := "127.0.0.1"
	port := strconv.Itoa(defaultPort)

	if string(target["enabled"]) == "1" && rawString(target["ip"]) == ip && rawString(target["port"]) == port {
		return nil
	}

	target["enabled"] = json.RawMessage("1")
	target["ip"] = json.RawMessage(strconv.Quote(ip))
	target["port"] = json.RawMessage(strconv.Quote(port))

	config["udp"], err = json.Marshal(targets)
	if err != nil {
		return err
	}
	updated, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}
	backupPath := telemetryBackupPath(path)
	if err := os.WriteFile(backupPath, original, 0600); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), "config.recovery-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	_, err = temp.Write(append(updated, '\n'))
	closeErr := temp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(temp.Name(), path)
}

func rawString(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(string(raw))
}

func telemetryBackupPath(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "config_backup.wreckscope.json")
}

func legacyTelemetryBackupPaths(configPath string) []string {
	dir := filepath.Dir(configPath)
	return []string{
		filepath.Join(dir, "config-backup.json"),
		filepath.Join(dir, "config-backup"),
	}
}

func findTelemetryBackup(configPath string) (string, error) {
	primary := telemetryBackupPath(configPath)
	if _, err := os.Stat(primary); err == nil {
		return primary, nil
	}
	for _, legacy := range legacyTelemetryBackupPaths(configPath) {
		if _, err := os.Stat(legacy); err == nil {
			return legacy, nil
		}
	}
	return "", fmt.Errorf("no config_backup.wreckscope.json found")
}

func BackupExists(configPath string) bool {
	_, err := findTelemetryBackup(configPath)
	return err == nil
}

func cleanupTelemetryBackups(configPath string) {
	_ = os.Remove(telemetryBackupPath(configPath))
	for _, legacy := range legacyTelemetryBackupPaths(configPath) {
		_ = os.Remove(legacy)
	}
}

// RestoreBackup writes sibling config_backup.wreckscope.json (or legacy names) over config.json, then deletes them.
func RestoreBackup(configPath string) error {
	configPath = strings.TrimSpace(configPath)
	if configPath == "" {
		return fmt.Errorf("telemetry config path required")
	}
	backupPath, err := findTelemetryBackup(configPath)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(configPath), "config.restore-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(temp.Name(), configPath); err != nil {
		return err
	}
	cleanupTelemetryBackups(configPath)
	return nil
}

// UpdateFile is the old name for EnableAt (tests still call it).
func UpdateFile(path string, port int) error {
	return EnableAt(path, port)
}
