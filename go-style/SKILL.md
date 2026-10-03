---
name: go-style
description: Personal conventions for writing and reviewing Go code. Use when writing, editing, refactoring, or reviewing Go source files (.go), go.mod, or Go tests in any project.
---

# Go Style

Personal conventions layered on standard Go idiom. Follow Effective Go and Go
Code Review Comments. These conventions refine that baseline; they never
override an established idiom.

## Tooling

Before committing, run:

- `go fmt ./...`
- `go vet ./...`
- `go fix ./...` to apply modernizations allowed by the `go.mod` Go version.
  Review the diff, since some fixes, such as `omitzero`, can change behavior.
- `go test ./...`

## Declaration order

Within a file, order top-level declarations as follows. Earlier rules win.

- Keep a type and its methods together, the type first.
- Group related declarations.
- Put package-level constants and variables first.
- Put exported declarations before unexported ones.
- Put declarations used in other files before those used only in this file.

## Errors

- A nil required argument is a programmer error: panic, do not return an
  error. Check explicitly at entry when the nil would otherwise surface
  later, away from the call site.

## Logging

- Log through `log/slog`. Third-party packages are fine only as
  `slog.Handler` implementations, never as a replacement logging API.

## CLI design

- Keep `main` minimal: it calls a testable `run` function and exits non-zero
  if `run` returns an error.
- `run` takes all process dependencies as arguments: a `context.Context`,
  args, stdin, stdout, stderr, and an environment lookup such as
  `os.Getenv`. It never reads `os` globals directly.
- `run` reports its own errors to stderr; `main` only sets the exit code.
- Test the CLI by calling `run` with in-memory readers and writers.

## Testing

- Test documented behavior, not implementation. A test should survive a
  refactor that preserves behavior.
- Avoid tautological tests, which only restate the code under test, such as
  deriving the expected value with the same logic.
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
- Mat Ryer. *How I write HTTP services in Go after 13 years*. Grafana Labs,
  2024. <https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/>
