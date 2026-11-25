package update

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"wreckscope/internal/version"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

const (
	githubRepo     = "lIMonkieIl/wreckscope-app"
	checksumAsset  = "SHA256SUMS"
	ModePortable   = "portable"
	ModeInstalled  = "installed"
)

// Status is returned to the UI after a check.
type Status struct {
	Current   string `json:"current"`
	Latest    string `json:"latest"`
	Available bool   `json:"available"`
	Mode      string `json:"mode"`
	Message   string `json:"message"`
}

// Configure wires app.Updater for this install mode. WindowNone — UI lives in Settings.
func Configure(app *application.App) error {
	portable := IsPortable()
	matcher := MatchSetup
	if portable {
		matcher = MatchPortable
	}
	gh, err := github.New(github.Config{
		Repository:    githubRepo,
		ChecksumAsset: checksumAsset,
		AssetMatcher:  matcher,
	})
	if err != nil {
		return err
	}
	return app.Updater.Init(updater.Config{
		CurrentVersion: version.Version,
		Providers:      []updater.Provider{gh},
		Window:         updater.WindowNone,
	})
}

// Mode returns portable or installed.
func Mode() string {
	if IsPortable() {
		return ModePortable
	}
	return ModeInstalled
}

// Check looks for a newer release without downloading.
func Check(ctx context.Context, app *application.App) (Status, error) {
	out := Status{
		Current: version.Version,
		Mode:    Mode(),
	}
	rel, err := app.Updater.Check(ctx)
	if err != nil {
		out.Message = err.Error()
		return out, err
	}
	if rel == nil {
		out.Message = "Up to date"
		return out, nil
	}
	out.Available = true
	out.Latest = rel.Version
	if out.Mode == ModePortable {
		out.Message = "Update " + rel.Version + " available"
	} else {
		out.Message = "Installer " + rel.Version + " available"
	}
	return out, nil
}

// Apply downloads the pending release. Portable: stage + restart (in-place swap).
// Installed: stage setup exe, launch it, then quit so the installer can replace files.
func Apply(ctx context.Context, app *application.App) (Status, error) {
	out := Status{
		Current: version.Version,
		Mode:    Mode(),
	}
	state := app.Updater.State()
	if state != updater.StateAvailable && state != updater.StateReady {
		st, err := Check(ctx, app)
		if err != nil {
			return st, err
		}
		if !st.Available {
			return st, nil
		}
		out = st
	}

	state = app.Updater.State()
	if state != updater.StateReady {
		if err := app.Updater.DownloadAndInstall(ctx); err != nil {
			out.Message = err.Error()
			return out, err
		}
	}
	out.Available = true

	if out.Mode == ModePortable {
		out.Message = "Restarting to apply update"
		if out.Latest != "" {
			out.Message = "Restarting to apply " + out.Latest
		}
		if err := app.Updater.Restart(ctx); err != nil {
			out.Message = err.Error()
			return out, err
		}
		return out, nil
	}

	path := app.Updater.DownloadedPath()
	if path == "" {
		out.Message = "Installer download missing"
		return out, fmt.Errorf("updater: empty downloaded path")
	}
	if err := launch(path); err != nil {
		out.Message = err.Error()
		return out, err
	}
	out.Message = "Opened installer"
	if out.Latest != "" {
		out.Message = "Opened installer " + out.Latest
	}
	go func() {
		time.Sleep(500 * time.Millisecond)
		app.Quit()
	}()
	return out, nil
}

func launch(path string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("cmd", "/C", "start", "", path).Start()
	default:
		cmd := exec.Command(path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Start()
	}
}
