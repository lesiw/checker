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
//
// # Ignore directives
//
// Diagnostics can be suppressed with comment directives. checker
// recognizes its own //ignore directive as well as the //nolint
// directive from golangci-lint and the //lint:ignore and
// //lint:file-ignore directives from staticcheck, so projects
// migrating from those tools can keep their existing comments.
//
//	//ignore:errcheck             // suppress errcheck on the next decl
//	//nolint:errcheck             // same, golangci-lint syntax
//	//lint:ignore SA1000 reason   // staticcheck syntax (reason required)
//	//lint:file-ignore U1000 generated code  // whole file
//
// An omitted analyzer list (//ignore or //nolint) or the special
// list "all" suppresses every analyzer. Multiple names are
// comma-separated. When a directive sits on its own line above a
// declaration it suppresses diagnostics for that declaration and
// its body; trailing on the same line as code it suppresses only
// that line. //lint:file-ignore applies to the whole file
// regardless of position.
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
