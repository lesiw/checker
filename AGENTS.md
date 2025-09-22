# AGENTS.md

## Operations

- Standard Go operations are available: `go run`, `go test`, `go build`,
  `go install`, `go generate`.
- Use `go test -race` to run tests with the race detector enabled.

## Style

Write Go in the spirit of _Effective Go_: clear but concise.

All code and documentation should be formatted to an 80-column limit.
Tabs are four characters.

### Comments & documentation

- Comments should only be used only when the code is not self-explanatory.
- It is preferable to make the code clearer than to explain it using a comment.
- Remove redundant comments that merely restate what the code does.
- Comments on a public (capitalized) name are part of user-facing documentation
  and must not be removed.
- If documentation is missing for a public name, add it.

### Implementation code

- Variable, function, and package names should be simple and evocative.
- Never name a package `pkg`.
- Blank lines in functions are discouraged except where adding them would make
  the code easier to understand.
- Implementation code should be as simple as possible, but no simpler.
- Unused and unnecessary code should be deleted.
- Use early returns to reduce nesting and avoid large if blocks.
- Do as much in a single line as possible where doing so does not harm
  readability or violate the 80 column limit.
- Combine variable declarations with conditions where possible using short
  variable declarations in if statements (e.g., `if x := getValue(); x != nil`).
- Use modern Go idioms. Prefer `any` over `interface{}`.
- A type that is a struct containing only one field should likely not be a
  struct. Promote the field to the top-level type.

When wrapping code to the 80 column limit, keep the opening and closing of
parenthesis or braces at the same indentation level. For example, a method
signature should be formatted like this:
```go
func fn(
	foo, bar, baz string
) error {
```
Not like this:
```
func fn(foo, bar,
	baz string) error {
```

For struct literals, function calls, and return statements, prioritize
single-line formatting when it doesn't violate the 80-column limit:
```go
return SomeStruct{value1, value2, value3}
```
Only use multi-line formatting when necessary, with consistent indentation and
trailing commas:
```go
return SomeStruct{
	Field1: value1,
	Field2: value2,
	Field3: value3,
}
```
Not like this:
```go
return SomeStruct{field1, field2,
	field3}
```

### Test code

- Do not remove existing tests without asking first.
- Unless already present, do not use assert libraries, such as testify's
  "assert."
- Prefer the "got, want" idiom.
- Prefer `t.Cleanup()` to `defer`.
- Identify the function that failed and its inputs when applicable.
  `t.Errorf("YourFunc(%v) = %v, want %v", in, got, want)`
- To compare complex structures, import github.com/google/go-cmp rather than
  using reflect.DeepEqual.
- Consider adding testable examples for public names.
- To make implementation code testable, use standard `io` interfaces where it
  makes sense to do so, such as `io.Reader` and `io.Writer`.
- Never use the `pkg_test` package unless it becomes absolutely necessary to do
  so to avoid an import loop.

When defining test tables, keep indentations to a minimum. Double up braces
when defining a list of test structs. Uncomfortably long lists of cases should
be broken out of test functions and into private global variables.
```go
var testCases = []struct{
	in   string
	want string
}{{
	"world", "Hello world",
}, {
	"foo", "Hello foo",
}}
```

Do not generate spurious interfaces for the sake of testability. Prefer test
hooks to dependency injection.
```go
var testHookGreet func(string) string

func Greet(s string) string {
    if h := testHookGreet; h != nil {
        return h(s)
    }
    return fmt.Sprintf("Hello, %s!", s)
}
```

The `swap` function is useful when testing code that modifies global values.
```go
func swap[T any](t *testing.T, orig *T, with T) {
    t.Helper()
    o := *orig
    t.Cleanup(func() { *orig = o })
    *orig = with
}
```

### When performing a final style pass

- Apply fixes from the "modernize" analyzer when making changes:
  `go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -fix -test ./...`
- Apply simplifications from gofmt when making changes: `gofmt -s -w ./...`
- Validate that the order of declarations makes sense. If someone were reading
  any individual source code file from top to bottom, is information presented
  in a way that is clear and easy to follow? Are related items grouped
  together?
