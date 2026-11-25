//go:build dev

package app

// buildAllowsDemo: true with `-tags dev` (e.g. wails3 task build DEV=true EXTRA_TAGS=dev).
func buildAllowsDemo() bool { return true }
