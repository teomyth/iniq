// Package version provides version information for the INIQ application
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

var (
	// These variables are set during build via -ldflags
	version   = "dev"
	buildDate = "unknown"
	commit    = "unknown"
)

// Info contains version information
type Info struct {
	// Version is the semantic version
	Version string
	// BuildDate is the date when the binary was built
	BuildDate string
	// Commit is the git commit hash
	Commit string
	// GoVersion is the version of Go used to build the binary
	GoVersion string
	// Platform is the OS/arch combination
	Platform string
	// BuildInfo contains additional build information
	BuildInfo *debug.BuildInfo
}

// Get returns version information
func Get() Info {
	info := Info{
		Version:   version,
		BuildDate: buildDate,
		Commit:    commit,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	// Try to get build info from runtime/debug
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		info.BuildInfo = buildInfo

		// If version is still "dev", try to get it from module info
		if info.Version == "dev" && buildInfo.Main.Version != "(devel)" && buildInfo.Main.Version != "" {
			info.Version = buildInfo.Main.Version
		}

		// Look for additional build settings
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				if info.Commit == "unknown" {
					info.Commit = setting.Value
				}
			case "vcs.time":
				if info.BuildDate == "unknown" {
					info.BuildDate = setting.Value
				}
			}
		}
	}

	return info
}

// String returns a string representation of version information
func String() string {
	info := Get()
	return fmt.Sprintf("%s (built: %s, commit: %s)", info.Version, info.BuildDate, info.Commit)
}

// ShortString returns a short string representation of version information
func ShortString() string {
	return Get().Version
}

// FormatBuildDate formats the build date as RFC3339
func FormatBuildDate(t time.Time) string {
	return t.Format(time.RFC3339)
}

// IsRelease returns true if the current version is a release version
func IsRelease() bool {
	v := Get().Version
	return v != "dev" && v != "unknown" && !strings.Contains(v, "-dev") && !strings.Contains(v, "-dirty")
}

// IsDevelopment returns true if the current version is a development version
func IsDevelopment() bool {
	return !IsRelease()
}
