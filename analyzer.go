package checker

import (
	"bytes"
	"testing"

	"golang.org/x/tools/go/analysis"
	check "golang.org/x/tools/go/analysis/checker"
	"golang.org/x/tools/go/packages"
)

// Run runs analyzers against the current package.
//
// If the analyzers produce diagnostics, or fail to run, the test will fail.
func Run(t *testing.T, analyzers ...*analysis.Analyzer) {
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: true,
	}, ".")
	if err != nil {
		t.Fatalf("[lesiw.io/checker] failed to load packages: %v", err)
	}
	graph, err := check.Analyze(analyzers, pkgs, nil)
	if err != nil {
		t.Fatalf("[lesiw.io/checker] failed to run analyzers: %v", err)
	}
	var buf bytes.Buffer
	if err := graph.PrintText(&buf, 0); err != nil {
		t.Fatalf("[lesiw.io/checker] failed to print diagnostics: %v", err)
	}
	if buf.Len() > 0 {
		t.Errorf("[lesiw.io/checker] check failed\n%v", buf.String())
	}
}
