// Package checker provides functions to run analyzers and linters as Go tests.
//
//	package main
//
//	import (
//	    "testing"
//
//	    "github.com/kisielk/errcheck/errcheck"
//	    "lesiw.io/checker"
//	)
//
//	func TestCheck(t *testing.T) {
//	    checker.Run(t, errcheck.Analyzer) // Run errcheck by itself.
//	    checker.Lint(t, "2.2.1")          // Run golangci-lint v2.2.1.
//	}
package checker

import (
	"os"
	"path/filepath"
)

func cacheDir(t T) string {
	cache, err := os.UserCacheDir()
	if err != nil {
		t.Fatalf("failed to get user cache directory: %v", err)
	}
	dir := filepath.Join(cache, "gochecker")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create cache directory: %v", err)
	}
	return dir
}
