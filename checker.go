// Package checker provides functions to run analyzers as Go tests.
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
//	    checker.Run(t, errcheck.Analyzer)
//	}
package checker

import (
	"bytes"
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

type testingT struct{ *testing.T }

func (t testingT) Fatalf(format string, args ...any) {
	t.T.Fatalf("[lesiw.io/checker] "+format, args...)
}

func (t testingT) Errorf(format string, args ...any) {
	t.T.Errorf("[lesiw.io/checker] "+format, args...)
}

func (t testingT) Logf(format string, args ...any) {
	t.T.Logf("[lesiw.io/checker] "+format, args...)
}
