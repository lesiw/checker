package checker

import (
	"testing"

	"github.com/kisielk/errcheck/errcheck"
	"lesiw.io/linelen"
	"lesiw.io/tidytypes"
)

func TestCheck(t *testing.T) {
	Run(t,
		errcheck.Analyzer,
		linelen.Analyzer,
		tidytypes.Analyzer,
	)
}
