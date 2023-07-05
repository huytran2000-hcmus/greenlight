package vcs

import (
	"fmt"
	"runtime/debug"
)

func Version() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	var revision string
	var modified bool
	for _, s := range buildInfo.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			if s.Value == "true" {
				modified = true
			}
		}
	}

	if modified {
		return fmt.Sprintf("%s-dirty", revision)
	}

	return revision
}
