package checker

import (
	"testing"

	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"lesiw.io/linelen"
	"lesiw.io/tidytypes"
)

func TestCheck(t *testing.T) {
	Run(t,
		errcheck.Analyzer,
		linelen.Analyzer,
		nilness.Analyzer,
		tidytypes.Analyzer,
	)
}
