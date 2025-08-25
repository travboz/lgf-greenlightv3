// // package vcs

// // import (
// // 	"fmt"
// // 	"runtime/debug"
// // )

// // func Version() string {
// // 	var revision string
// // 	var modified bool

// // 	bi, ok := debug.ReadBuildInfo()
// // 	if ok {
// // 		for _, s := range bi.Settings {
// // 			switch s.Key {
// // 			case "vcs.revision":
// // 				revision = s.Value
// // 			case "vcs.modified":
// // 				if s.Value == "true" {
// // 					modified = true
// // 				}
// // 			}
// // 		}
// // 	}

// // 	if modified {
// // 		// this just checks if the version has been modified and is now considered "dirty" because it doesn't actually match the hash we'd expect
// // 		return fmt.Sprintf("%s-dirty", revision)
// // 	}

// // 	return revision
// // }

// package vcs

// import (
// 	"fmt"
// 	"runtime/debug"
// )

// var revision = "" // gets overridden via -ldflags

// func Version() string {
// 	if revision != "" {
// 		return revision
// 	}

// 	var rev string
// 	var modified bool
// 	if bi, ok := debug.ReadBuildInfo(); ok {
// 		for _, s := range bi.Settings {
// 			switch s.Key {
// 			case "vcs.revision":
// 				rev = s.Value
// 			case "vcs.modified":
// 				if s.Value == "true" {
// 					modified = true
// 				}
// 			}
// 		}
// 	}

// 	if rev == "" {
// 		return "unknown"
// 	}
// 	if modified {
// 		return fmt.Sprintf("%s-dirty", rev)
// 	}
// 	return rev
// }

package vcs

import (
	"fmt"
	"runtime/debug"
)

var (
	revision = "" // commit hash, overridden via -ldflags
	dirty    = "" // "true" if dirty, also overridden via -ldflags
)

func Version() string {
	// Case 1: injected at build time via -ldflags
	if revision != "" {
		if dirty == "true" {
			return fmt.Sprintf("%s-dirty", revision)
		}
		return revision
	}

	// Case 2: fallback to Go's debug.ReadBuildInfo
	var rev string
	var modified bool
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				rev = s.Value
			case "vcs.modified":
				if s.Value == "true" {
					modified = true
				}
			}
		}
	}

	if rev == "" {
		return "unknown"
	}
	if modified {
		return fmt.Sprintf("%s-dirty", rev)
	}
	return rev
}
