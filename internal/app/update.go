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

// UpdateMode is "installed" (NSIS + AppData).
func (a *App) UpdateMode() string {
	return update.Mode()
}

func (a *App) storeUpdateStatus(st update.Status) {
	a.updateMu.Lock()
	a.lastUpdate = st
	a.updateMu.Unlock()
}

func (a *App) loadUpdateStatus() update.Status {
	a.updateMu.Lock()
	defer a.updateMu.Unlock()
	return a.lastUpdate
}

func (a *App) emitUpdateChecking(active bool) {
	if a.wails == nil {
		return
	}
	a.wails.Event.Emit("update-checking", active)
}

// CheckForUpdates queries GitHub Releases. Does not download.
func (a *App) CheckForUpdates() (update.Status, error) {
	if a.wails == nil {
		return update.Status{Current: version.Version, Message: "app not ready"}, errNoAppContext
	}
	a.emitUpdateChecking(true)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	st, err := update.Check(ctx, a.wails)
	if err == nil {
		a.storeUpdateStatus(st)
		a.wails.Event.Emit("update-status", st)
		if st.Available {
			a.wails.Event.Emit("update-available", st)
		}
	} else {
		a.emitUpdateChecking(false)
	}
	return st, err
}

// ApplyUpdate downloads the setup installer to AppData, launches it, then quits.
func (a *App) ApplyUpdate() (update.Status, error) {
	if a.wails == nil {
		return update.Status{Current: version.Version, Message: "app not ready"}, errNoAppContext
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	st, err := update.Apply(ctx, a.wails)
	prev := a.loadUpdateStatus()
	if st.Latest == "" && prev.Latest != "" {
		st.Latest = prev.Latest
		st.Available = true
	}
	if err == nil {
		a.storeUpdateStatus(st)
	}
	return st, err
}

// LastUpdateStatus returns the most recent check/apply result (may be empty).
func (a *App) LastUpdateStatus() update.Status {
	return a.loadUpdateStatus()
}

func (a *App) quietUpdateCheck() {
	if a.wails == nil {
		return
	}
	// Initial check shortly after start, then periodically while running.
	select {
	case <-a.ctx.Done():
		return
	case <-time.After(4 * time.Second):
	}
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
	if a.wails == nil || a.ctx == nil || a.ctx.Err() != nil {
		return
	}
	a.emitUpdateChecking(true)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	st, err := update.Check(ctx, a.wails)
	if err != nil {
		log.Printf("Quiet update check: %v", err)
		a.emitUpdateChecking(false)
		return
	}
	if a.ctx.Err() != nil {
		a.emitUpdateChecking(false)
		return
	}
	a.storeUpdateStatus(st)
	a.wails.Event.Emit("update-status", st)
	if st.Available {
		a.wails.Event.Emit("update-available", st)
	}
}
