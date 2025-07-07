# lesiw.io/checker [![Go Reference](https://pkg.go.dev/badge/lesiw.io/checker.svg)](https://pkg.go.dev/lesiw.io/checker)

Package checker provides functions to run analyzers and linters as Go tests.

```go
package main

import (
    "testing"

    "github.com/kisielk/errcheck/errcheck"
    "lesiw.io/checker"
)

func TestCheck(t *testing.T) {
    checker.Run(t, errcheck.Analyzer) // Run errcheck by itself.
    checker.Lint(t, "2.2.1")          // Run golangci-lint v2.2.1.
}
```

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
