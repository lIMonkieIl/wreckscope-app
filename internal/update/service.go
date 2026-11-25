package update

import (
	"context"
	"os"
	"os/exec"
	"runtime"

	"wreckscope/internal/version"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

const (
	githubRepo    = "lIMonkieIl/wreckscope-app"
	checksumAsset = "SHA256SUMS"
	ModeInstalled = "installed"
)

// Status is returned to the UI after a check.
type Status struct {
	Current   string `json:"current"`
	Latest    string `json:"latest"`
	Available bool   `json:"available"`
	Mode      string `json:"mode"`
	Message   string `json:"message"`
}

// Configure wires app.Updater for installed builds. WindowNone — UI lives in Settings.
func Configure(app *application.App) error {
	gh, err := github.New(github.Config{
		Repository:    githubRepo,
		ChecksumAsset: checksumAsset,
		AssetMatcher:  MatchSetup,
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

// Mode is always installed (NSIS + AppData).
func Mode() string {
	return ModeInstalled
}

// Check looks for a newer release without downloading.
func Check(ctx context.Context, app *application.App) (Status, error) {
	out := Status{
		Current: version.Version,
		Mode:    ModeInstalled,
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
	out.Message = "Installer " + rel.Version + " available"
	return out, nil
}

// Apply downloads the setup package to %AppData%\WreckScope\updates, launches it,
// then quits. Does not stage beside Program Files (that caused the Download installer no-op).
func Apply(ctx context.Context, app *application.App) (Status, error) {
	return applyInstalled(ctx, app, Status{
		Current: version.Version,
		Mode:    ModeInstalled,
	})
}

func launch(path string) error {
	switch runtime.GOOS {
	case "windows":
		// start "" <path> — empty title so path is not treated as the window title.
		return exec.Command("cmd", "/C", "start", "", path).Start()
	default:
		cmd := exec.Command(path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Start()
	}
}
