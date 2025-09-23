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
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis"
	gochecker "golang.org/x/tools/go/analysis/checker"
	"golang.org/x/tools/go/packages"
)

// Run runs analyzers against the current package.
//
// If the analyzers produce diagnostics, or fail to run, the test will fail.
func Run(t *testing.T, analyzers ...*analysis.Analyzer) {
	run(testingT{t}, analyzers...)
}

func run(t testingT, analyzers ...*analysis.Analyzer) {
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: true,
	}, ".")
	if err != nil {
		t.Fatalf("failed to load packages: %v", err)
	}

	graph, err := gochecker.Analyze(
		[]*analysis.Analyzer{NewAnalyzer(analyzers...)}, pkgs, nil,
	)
	if err != nil {
		t.Fatalf("failed to run analyzers: %v", err)
	}

	var buf bytes.Buffer
	if err := graph.PrintText(&buf, 0); err != nil {
		t.Fatalf("failed to print diagnostics: %v", err)
	}
	if buf.Len() > 0 {
		t.Errorf("check failed\n%v", buf.String())
	}
}

func cacheDir(t testingT) string {
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
