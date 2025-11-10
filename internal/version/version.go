package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current version of the application
	// This will be set at build time using -ldflags
	Version = "dev"

	// GitCommit is the git commit hash
	// This will be set at build time using -ldflags
	GitCommit = "unknown"

	// BuildDate is the date when the binary was built
	// This will be set at build time using -ldflags
	BuildDate = "unknown"
)

// Info contains version information
type Info struct {
	Version   string
	GitCommit string
	BuildDate string
	GoVersion string
	Platform  string
	Arch      string
}

// GetInfo returns the version information
func GetInfo() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf(`Todo-Go Version Information:
  Version:    %s
  Git Commit: %s
  Build Date: %s
  Go Version: %s
  Platform:   %s/%s`,
		i.Version,
		i.GitCommit,
		i.BuildDate,
		i.GoVersion,
		i.Platform,
		i.Arch,
	)
}

// Short returns a short version string
func (i Info) Short() string {
	return fmt.Sprintf("v%s (%s/%s)", i.Version, i.Platform, i.Arch)
}
