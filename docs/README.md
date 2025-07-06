# checker

Package checker provides functions to run analyzers and linters as Go tests.

```go
package main

import (
    "testing"

    "github.com/kisielk/errcheck/errcheck"
    "lesiw.io/checker"
)

func TestCheck(t *testing.T) {
    checker.Run(t, errcheck.Analyzer)
}
```

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
