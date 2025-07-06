package main

import (
	"fmt"
	"os"
	"strings"

	"labs.lesiw.io/ops/golang"
	"labs.lesiw.io/ops/golib"
	"lesiw.io/cmdio"
	"lesiw.io/cmdio/sys"
	"lesiw.io/ops"
)

type Ops struct{ golib.Ops }

func main() {
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "check")
	}
	ops.Handle(Ops{})
}

func (o Ops) Doc() error {
	rnr := golang.Source().WithCommand("go", sys.Runner())
	r, err := rnr.Get("go", "tool", "goreadme", "-skip-sub-packages")
	if err != nil {
		return fmt.Errorf("goreadme failed: %w", err)
	}
	_, err = cmdio.GetPipe(
		strings.NewReader(r.Out+"\n"),
		rnr.Command("tee", "docs/README.md"),
	)
	if err != nil {
		return fmt.Errorf("could not update README.md: %w", err)
	}
	return nil
}
