package version

import (
	"fmt"
	"runtime"
)

var (
	MajorVersion = "beta"
	GitHash      = "Unknown"
	GoVer        = runtime.Version()
)

func Version() string {
	return fmt.Sprintf("Version: %s\nGitHash: %s\nGoVer: %s\n", MajorVersion, GitHash, GoVer)
}
