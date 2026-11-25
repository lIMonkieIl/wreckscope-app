package app

import "errors"

var (
	errInvalidMode          = errors.New("invalid mode")
	errInvalidTelemetryPath = errors.New("telemetry config path required")
	errNoAppContext         = errors.New("app not ready")
)
