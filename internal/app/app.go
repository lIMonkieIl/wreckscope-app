package app

import (
	"context"
	"embed"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"wreckscope/internal/racefeed"
	"wreckscope/internal/settings"
	"wreckscope/internal/setup"
	"wreckscope/internal/update"
	"wreckscope/internal/wf2"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const defaultUDPPort = 23123
const defaultHTTPPort = 8080

// App is the Wails backend service.
type App struct {
	ctx      context.Context
	wails    *application.App
	engine   *racefeed.Engine
	signal   chan struct{}
	udpPort  int
	httpPort int
	lan      bool
	assets   embed.FS

	httpServer *http.Server
	cancel     context.CancelFunc
	wg         sync.WaitGroup

	listenMu   sync.Mutex
	httpReady  bool
	udpReady   bool
	httpBindErr string
	udpBindErr  string

	updateMu   sync.Mutex
	lastUpdate update.Status
}

func New(assets embed.FS) *App {
	return &App{
		engine:   racefeed.New(),
		signal:   make(chan struct{}, 1),
		udpPort:  defaultUDPPort,
		httpPort: defaultHTTPPort,
		lan:      true, // LAN on by default
		assets:   assets,
	}
}

func (a *App) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	a.wails = application.Get()
	runCtx, cancel := context.WithCancel(ctx)
	a.ctx = runCtx
	a.cancel = cancel

	path, err := settings.Path()
	if err != nil {
		log.Printf("Find settings file: %v", err)
	} else if saved, err := settings.Load(path); err == nil {
		_, _ = a.engine.ApplySettings(saved, nil)
	} else if os.IsNotExist(err) {
		// First run: seed AppData with brand defaults (wreckscope + system).
		if saveErr := settings.Save(path, a.engine.Settings()); saveErr != nil {
			log.Printf("Seed settings: %v", saveErr)
		}
	} else {
		log.Printf("Load settings: %v", err)
	}
	a.migrateLegacyTelemetryPath()
	if err := update.ClearStaged(); err != nil {
		log.Printf("Clear staged updates: %v", err)
	}

	a.wg.Add(4)
	go func() {
		defer a.wg.Done()
		a.listenUDP(runCtx)
	}()
	go func() {
		defer a.wg.Done()
		a.demoLoop(runCtx)
	}()
	go func() {
		defer a.wg.Done()
		a.emitLoop(runCtx)
	}()
	go func() {
		defer a.wg.Done()
		a.startHTTP(runCtx)
	}()
	go a.quietUpdateCheck()
	return nil
}

func (a *App) ServiceShutdown() error {
	if a.cancel != nil {
		a.cancel()
	}
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
	return nil
}

func (a *App) notify() {
	select {
	case a.signal <- struct{}{}:
	default:
	}
}

func (a *App) setUDPListen(ok bool, errMsg string) {
	a.listenMu.Lock()
	a.udpReady = ok
	a.udpBindErr = errMsg
	a.listenMu.Unlock()
	a.refreshBindMessage()
}

func (a *App) setHTTPListen(ok bool, errMsg string) {
	a.listenMu.Lock()
	a.httpReady = ok
	a.httpBindErr = errMsg
	a.listenMu.Unlock()
	a.refreshBindMessage()
}

func (a *App) httpListening() bool {
	a.listenMu.Lock()
	defer a.listenMu.Unlock()
	return a.httpReady
}

func (a *App) refreshBindMessage() {
	a.listenMu.Lock()
	udpErr, httpErr := a.udpBindErr, a.httpBindErr
	a.listenMu.Unlock()
	switch {
	case udpErr != "" && httpErr != "":
		a.engine.SetMessage(udpErr + " · " + httpErr)
	case udpErr != "":
		a.engine.SetMessage(udpErr)
	case httpErr != "":
		a.engine.SetMessage(httpErr)
	default:
		a.engine.SetMessage("")
	}
	a.notify()
}

func (a *App) listenUDP(ctx context.Context) {
	sock, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: a.udpPort})
	if err != nil {
		msg := "UDP " + strconv.Itoa(a.udpPort) + " in use — close the other recovery tool"
		log.Printf("UDP bind failed: %v", err)
		a.setUDPListen(false, msg)
		return
	}
	defer sock.Close()
	a.setUDPListen(true, "")
	log.Printf("UDP listening 127.0.0.1:%d", a.udpPort)
	_ = sock.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buf := make([]byte, 65535)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_ = sock.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _, err := sock.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			log.Printf("UDP error: %v", err)
			continue
		}
		if a.engine.Mode() != "live" {
			continue
		}
		if wf2.Decode(buf[:n], a.engine) {
			a.notify()
		}
	}
}

func (a *App) demoLoop(ctx context.Context) {
	if !a.demoFeaturesEnabled() {
		if a.engine.Mode() == "demo" {
			a.engine.SetMode("live")
			a.notify()
		}
		return
	}
	t := 0.0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if a.engine.Mode() == "demo" {
				a.engine.DemoTick(t)
				a.notify()
				t += 0.1
			} else {
				t = 0
			}
		}
	}
}

func (a *App) emitLoop(ctx context.Context) {
	frame := time.NewTicker(33 * time.Millisecond)
	refresh := time.NewTicker(100 * time.Millisecond)
	defer frame.Stop()
	defer refresh.Stop()
	dirty := true
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.signal:
			dirty = true
		case <-refresh.C:
			dirty = true
		case <-frame.C:
			if !dirty || a.wails == nil {
				continue
			}
			dirty = false
			a.wails.Event.Emit("telemetry", a.engine.Snapshot())
		}
	}
}

func (a *App) Snapshot() racefeed.Snapshot {
	return a.engine.Snapshot()
}

// ToggleMedic assigns or clears a medic slot (PC + LAN; no cap).
func (a *App) ToggleMedic(slot int) []int {
	medics := a.engine.ToggleMedic(slot)
	a.notify()
	return medics
}

// demoFeaturesEnabled: `dev` build tag or non-production (debug) Wails build. Off in release.
func (a *App) demoFeaturesEnabled() bool {
	if buildAllowsDemo() {
		return true
	}
	if a.wails == nil {
		return false
	}
	return a.wails.Env.Info().Debug
}

// SetMode switches live/demo. Demo rejected outside development builds.
func (a *App) SetMode(mode string) error {
	if mode != "demo" && mode != "live" {
		return errInvalidMode
	}
	if mode == "demo" && !a.demoFeaturesEnabled() {
		return errInvalidMode
	}
	a.engine.SetMode(mode)
	a.notify()
	return nil
}

func (a *App) GetSettings() racefeed.Settings {
	return a.engine.Settings()
}

// SaveSettings persists shared settings. Empty telemetryConfigPath / setupCompleted
// keep current values (UI forms omit them).
func (a *App) SaveSettings(settings racefeed.Settings) (racefeed.Settings, error) {
	current := a.engine.Settings()
	if strings.TrimSpace(settings.TelemetryConfigPath) == "" {
		settings.TelemetryConfigPath = current.TelemetryConfigPath
	}
	settings.SetupCompleted = current.SetupCompleted
	return a.persistSettings(settings)
}

func (a *App) SaveAppearance(theme, mode string) (racefeed.Settings, error) {
	current := a.engine.Settings()
	current.Theme = theme
	current.Mode = mode
	return a.persistSettings(current)
}

// SetupStatus: first-run + telemetry probe. needsWizard == !setupCompleted.
func (a *App) SetupStatus() map[string]interface{} {
	status := a.ProbeTelemetry()
	completed := a.engine.Settings().SetupCompleted
	status["setupCompleted"] = completed
	status["needsWizard"] = !completed
	return status
}

func (a *App) CompleteSetup() (map[string]interface{}, error) {
	current := a.engine.Settings()
	current.SetupCompleted = true
	if _, err := a.persistSettings(current); err != nil {
		return a.SetupStatus(), err
	}
	return a.SetupStatus(), nil
}

func (a *App) emitSettings(settings racefeed.Settings) {
	if a.wails == nil {
		return
	}
	a.wails.Event.Emit("settings", settings)
}

// PhoneAccess: LAN phone URLs for the Device view QR. Full app, not view-only.
func (a *App) PhoneAccess() map[string]interface{} {
	addresses := []string{}
	listening := a.lan && a.httpListening()
	if listening {
		if address := setup.PhoneURL(a.httpPort); address != "" {
			addresses = append(addresses, address)
		}
	}
	out := map[string]interface{}{"lan": a.lan, "urls": addresses, "readOnly": false, "listening": listening}
	a.listenMu.Lock()
	httpErr := a.httpBindErr
	a.listenMu.Unlock()
	if a.lan && !listening {
		if httpErr != "" {
			out["error"] = httpErr
		} else {
			out["error"] = "LAN dashboard is not listening"
		}
	} else if listening && len(addresses) == 0 {
		out["error"] = "No primary private IPv4 address found."
	}
	return out
}


func (a *App) TelemetryStatus() map[string]interface{} {
	return a.ProbeTelemetry()
}

func (a *App) telemetryStatus(path string, canAuto bool, errMsg string) map[string]interface{} {
	path = strings.TrimSpace(path)
	var validErr error
	if path != "" {
		validErr = setup.LooksLikeConfig(path)
	}
	valid := path != "" && validErr == nil
	ready := valid && setup.Ready(path, a.udpPort)
	if errMsg == "" && validErr != nil {
		errMsg = validErr.Error()
	}
	completed := a.engine.Settings().SetupCompleted
	return map[string]interface{}{
		"ready":          ready,
		"valid":          valid,
		"setupCompleted": completed,
		"needsWizard":    !completed,
		"canAuto":        canAuto,
		"hasBackup":      valid && setup.BackupExists(path),
		"path":           path,
		"error":          errMsg,
	}
}

// ProbeTelemetry reads the game config; does not require it already enabled.
func (a *App) ProbeTelemetry() map[string]interface{} {
	path, err := a.resolveTelemetryPath()
	if err != nil {
		_, autoErr := setup.ConfigPath()
		return a.telemetryStatus("", autoErr == nil, err.Error())
	}
	_, autoErr := setup.ConfigPath()
	return a.telemetryStatus(path, autoErr == nil, "")
}

// ProbeTelemetryAt checks a chosen config.json path without writing.
func (a *App) ProbeTelemetryAt(path string) map[string]interface{} {
	path = strings.TrimSpace(path)
	_, autoErr := setup.ConfigPath()
	if path == "" {
		return a.telemetryStatus("", autoErr == nil, "choose the game telemetry config.json")
	}
	return a.telemetryStatus(path, autoErr == nil, "")
}

// FindTelemetryConfig locates the game telemetry config.json (no write).
func (a *App) FindTelemetryConfig() map[string]interface{} {
	path, err := setup.ConfigPath()
	if err != nil {
		_, autoErr := setup.ConfigPath()
		return a.telemetryStatus("", autoErr == nil, err.Error())
	}
	return a.telemetryStatus(path, true, "")
}

// EnableTelemetry enables UDP on the auto-detected config path.
func (a *App) EnableTelemetry() (map[string]interface{}, error) {
	path, err := setup.ConfigPath()
	if err != nil {
		return a.telemetryStatus("", false, err.Error()), err
	}
	return a.EnableTelemetryAt(path)
}

// EnableTelemetryAt enables UDP on path and remembers it.
func (a *App) EnableTelemetryAt(path string) (map[string]interface{}, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errInvalidTelemetryPath
	}
	if err := setup.LooksLikeConfig(path); err != nil {
		return a.telemetryStatus(path, true, err.Error()), err
	}
	if setup.Ready(path, a.udpPort) {
		_ = a.rememberTelemetryPath(path)
		return a.telemetryStatus(path, true, ""), nil
	}
	if err := setup.EnableAt(path, a.udpPort); err != nil {
		return a.telemetryStatus(path, true, err.Error()), err
	}
	_ = a.rememberTelemetryPath(path)
	return a.telemetryStatus(path, true, ""), nil
}

// RestoreTelemetryBackup writes config_backup.wreckscope.json over config.json, then deletes the backup.
func (a *App) RestoreTelemetryBackup(path string) (map[string]interface{}, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = a.savedTelemetryPath()
	}
	if path == "" {
		resolved, err := a.resolveTelemetryPath()
		if err != nil {
			return a.telemetryStatus("", false, err.Error()), err
		}
		path = resolved
	}
	if err := setup.RestoreBackup(path); err != nil {
		return a.telemetryStatus(path, true, err.Error()), err
	}
	_ = a.rememberTelemetryPath(path)
	return a.telemetryStatus(path, true, ""), nil
}

// ResetTelemetrySetup clears the saved telemetry path. Leaves setupCompleted alone.
func (a *App) ResetTelemetrySetup() (map[string]interface{}, error) {
	if err := a.clearSavedTelemetryPath(); err != nil {
		return nil, err
	}
	status := a.ProbeTelemetry()
	status["ready"] = false
	return status, nil
}

func (a *App) BrowseTelemetryConfig() (string, error) {
	if a.wails == nil {
		return "", errNoAppContext
	}
	return a.wails.Dialog.OpenFile().
		SetTitle("Telemetry config.json").
		AddFilter("JSON (*.json)", "*.json").
		PromptForSingleSelection()
}
