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
	golib.Targets = []golib.Target{
		{Goos: "linux", Goarch: "386"},
		{Goos: "linux", Goarch: "amd64"},
		{Goos: "linux", Goarch: "arm"},
		{Goos: "linux", Goarch: "arm64"},
		{Goos: "darwin", Goarch: "amd64"},
		{Goos: "darwin", Goarch: "arm64"},
		{Goos: "windows", Goarch: "386"},
		{Goos: "windows", Goarch: "arm"},
		{Goos: "windows", Goarch: "amd64"},
	}
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
	content := "# lesiw.io/checker " +
		"[![Go Reference](https://pkg.go.dev/badge/lesiw.io/checker.svg)]" +
		"(https://pkg.go.dev/lesiw.io/checker)\n" +
		r.Out[strings.Index(r.Out, "\n")+1:] + "\n"
	_, err = cmdio.GetPipe(
		strings.NewReader(content),
		rnr.Command("tee", "docs/README.md"),
	)
	if err != nil {
		return fmt.Errorf("could not update README.md: %w", err)
	}
	return nil
}
