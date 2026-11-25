package update

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"wreckscope/internal/settings"
	"wreckscope/internal/version"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

// applyInstalled downloads the setup package under %AppData%\WreckScope\updates,
// launches it, then quits. Wails DownloadAndInstall stages beside the running
// exe (Program Files) and fails without elevation — that was the no-op.
func applyInstalled(ctx context.Context, app *application.App, out Status) (Status, error) {
	gh, err := github.New(github.Config{
		Repository:    githubRepo,
		ChecksumAsset: checksumAsset,
		AssetMatcher:  MatchSetup,
	})
	if err != nil {
		out.Message = err.Error()
		return out, err
	}

	rel, err := gh.Check(ctx, updater.CheckRequest{
		CurrentVersion: version.Version,
		Platform:       runtime.GOOS,
		Arch:           runtime.GOARCH,
	})
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
	out.Message = "Downloading installer " + rel.Version

	dest, err := downloadSetup(ctx, gh, rel)
	if err != nil {
		out.Message = err.Error()
		return out, err
	}

	if err := launch(dest); err != nil {
		out.Message = err.Error()
		return out, err
	}

	out.Message = "Opened installer " + rel.Version
	go func() {
		time.Sleep(500 * time.Millisecond)
		app.Quit()
	}()
	return out, nil
}

func downloadSetup(ctx context.Context, gh *github.Provider, rel *updater.Release) (string, error) {
	dir, err := settings.UpdatesDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("updater: create updates dir: %w", err)
	}
	// Drop leftover staged installers before a new download.
	_ = clearDirFiles(dir)

	name := filepath.Base(rel.Artifact.Filename)
	if name == "" || name == "." || name == string(filepath.Separator) {
		name = "WreckScope-setup.exe"
	}
	dest := filepath.Join(dir, name)
	tmp := dest + ".partial"
	_ = os.Remove(tmp)

	f, err := os.Create(tmp)
	if err != nil {
		return "", fmt.Errorf("updater: create download: %w", err)
	}

	hasher := sha256.New()
	w := io.MultiWriter(f, hasher)
	dlErr := gh.Download(ctx, rel, w, nil)
	closeErr := f.Close()
	if dlErr != nil {
		_ = os.Remove(tmp)
		return "", dlErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return "", closeErr
	}

	sum := hasher.Sum(nil)
	if err := requireSHA256(rel, sum, name); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}

	_ = os.Remove(dest)
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("updater: finalize download: %w", err)
	}
	return dest, nil
}
