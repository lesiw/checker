package checker

import "testing"

type T struct{ *testing.T }

func (t T) Fatalf(format string, args ...any) {
	t.T.Fatalf("[lesiw.io/checker] "+format, args...)
}

func (t T) Errorf(format string, args ...any) {
	t.T.Errorf("[lesiw.io/checker] "+format, args...)
}

func (t T) Logf(format string, args ...any) {
	t.T.Logf("[lesiw.io/checker] "+format, args...)
}
