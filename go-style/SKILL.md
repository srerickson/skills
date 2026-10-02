---
name: go-style
description: Personal conventions for writing and reviewing Go code. Use when writing, editing, refactoring, or reviewing Go source files (.go), go.mod, or Go tests in any project.
---

# Go Style

Personal conventions layered on standard Go idiom. Format with `gofmt` and
follow Effective Go and Go Code Review Comments. These conventions refine that
baseline; they never override an established idiom.

## Declaration order

Within a file, order top-level declarations as follows. Earlier rules win.

- Keep a type and its methods together, the type first.
- Group related declarations.
- Put package-level constants and variables first.
- Put exported declarations before unexported ones.
- Put declarations used in other files before those used only in this file.

## Errors

- TODO

## Testing

- Each `foo_test.go` tests the code in `foo.go`. Do not create one-off test
  files.
- If it is unclear which file a test belongs in, that is a smell: the code
  under test may be misplaced, a file may need splitting, or the test is an
  integration test.
- Package-level test files are exempt: `example_test.go`, `export_test.go`,
  `main_test.go` (for `TestMain`), and shared test helpers.
- Integration and end-to-end tests are exempt. Keep them apart from unit tests
  with a separate directory, a build tag, or `testing.Short()`.

## References

- The Go Authors. *Effective Go*. <https://go.dev/doc/effective_go>
- The Go Authors. *Go Code Review Comments*. <https://go.dev/wiki/CodeReviewComments>
