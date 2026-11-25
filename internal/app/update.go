package app

import (
	"context"
	"log"
	"time"

	"wreckscope/internal/update"
	"wreckscope/internal/version"
)

// AppVersion is the bake-time version string.
func (a *App) AppVersion() string {
	return version.Version
}

// UpdateMode is "portable" or "installed".
func (a *App) UpdateMode() string {
	return update.Mode()
}

// CheckForUpdates queries GitHub Releases. Does not download.
func (a *App) CheckForUpdates() (update.Status, error) {
	if a.wails == nil {
		return update.Status{Current: version.Version, Message: "app not ready"}, errNoAppContext
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	st, err := update.Check(ctx, a.wails)
	if err == nil {
		a.lastUpdate = st
		a.wails.Event.Emit("update-status", st)
		if st.Available {
			a.wails.Event.Emit("update-available", st)
		}
	}
	return st, err
}

// ApplyUpdate downloads and applies the pending update (call CheckForUpdates first).
// Portable: in-place swap + restart. Installed: launch setup installer.
func (a *App) ApplyUpdate() (update.Status, error) {
	if a.wails == nil {
		return update.Status{Current: version.Version, Message: "app not ready"}, errNoAppContext
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	st, err := update.Apply(ctx, a.wails)
	if st.Latest == "" && a.lastUpdate.Latest != "" {
		st.Latest = a.lastUpdate.Latest
		st.Available = true
	}
	if err == nil {
		a.lastUpdate = st
	}
	return st, err
}

// LastUpdateStatus returns the most recent check/apply result (may be empty).
func (a *App) LastUpdateStatus() update.Status {
	return a.lastUpdate
}

func (a *App) quietUpdateCheck() {
	if a.wails == nil {
		return
	}
	// Initial check shortly after start, then periodically while running.
	time.Sleep(4 * time.Second)
	a.runQuietCheck()
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.runQuietCheck()
		}
	}
}

func (a *App) runQuietCheck() {
	if a.wails == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	st, err := update.Check(ctx, a.wails)
	if err != nil {
		log.Printf("Quiet update check: %v", err)
		return
	}
	a.lastUpdate = st
	a.wails.Event.Emit("update-status", st)
	if st.Available {
		a.wails.Event.Emit("update-available", st)
	}
}
