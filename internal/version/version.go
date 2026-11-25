// Package version holds the app version baked in at link time.
package version

// Version is set via -ldflags "-X wreckscope/internal/version.Version=…".
// Keep the default in sync with build/config.yml info.version.
var Version = "0.1.0"
