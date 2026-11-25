//go:build dev

package app

// buildAllowsDemo: true with `-tags dev` (e.g. wails dev -tags dev).
func buildAllowsDemo() bool { return true }
