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

import "testing"

func fatal(t *testing.T, format string, a ...any) {
	t.Fatalf("[lesiw.io/checker] "+format, a...)
}
