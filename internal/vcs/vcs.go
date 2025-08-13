package vcs

import (
	"fmt"
	"runtime/debug"
)

func Version() string {
	var revision string
	var modified bool

	bi, ok := debug.ReadBuildInfo()
	if ok {
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				revision = s.Value
			case "vcs.modified":
				if s.Value == "true" {
					modified = true
				}
			}
		}
	}

	if modified {
		// this just checks if the version has been modified and is now considered "dirty" because it doesn't actually match the hash we'd expect
		return fmt.Sprintf("%s-dirty", revision)
	}

	return revision
}
