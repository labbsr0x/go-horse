package config

// Default build-time variable.
// These values are overridden via ldflags
var (
	Version   = "unknown-version"
	GitCommit = "unknown-commit"
	BuildTime = "unknown-buildtime"
)
