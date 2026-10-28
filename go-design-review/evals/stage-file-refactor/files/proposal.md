# Proposal: move the stage file lifecycle into the stage package

Refactor of the stage cache work (see `issue.md`). The current code is in
`ocfl-tools/` (ocfl-go v0.13.0 is in the Go module cache).

## Where the boundary leaks today

- `run` keeps the stage file in sync by hand. It calls `stage.Write` after
  every add and rm, `os.Remove(ref.path)` after a commit and in `delete`,
  `stage.Write(ref.path, draft)` on a conflict, and `stage.Write` after a
  revert and after `status` refreshes stats.
- `applyStage(ctx, env, root, s, ref, draft, resuming, content)` mixes
  persistence, the commit lifecycle, and presentation. `StageCommitCmd.Run`
  reads the stage file a second time just to keep a pre-finalize copy as
  `draft`; `ocfl commit` passes `nil`. A draft and a frozen stage differ only
  in the version metadata (message, user, created). In ocfl-go v0.13.0,
  `Finalize` does no I/O, and `Revert` turns the update back into a draft by
  clearing it, but there's no way to undo a finalize without I/O.
- `resuming bool` stands in for "which lifecycle step are we in".
- `errStagedContent` is domain knowledge in `run`; it exists because
  `Content` takes an `open` callback that `run` wraps.

## Proposal: the stage is the domain object, the file is how it's stored

`Stage` stays a draft or frozen update plus records of staged files. A new
`stage.File` is a stage bound to its file. Any method that moves the stage
to a new state saves the file before returning, so `run` never writes,
removes, or re-reads stage files.

```go
// package stage

func Open(path string) (*File, error)              // replaces Read
func Create(path string, s *Stage) (*File, error)  // replaces WriteNew; fails with fs.ErrExist

func (f *File) Stage() *Stage
func (f *File) Save() error                         // after editing a draft (add, rm)
func (f *File) Check(ctx) (*CheckResult, error)     // saves refreshed stats itself
func (f *File) Delete() error                       // refuses a frozen stage...
func (f *File) Discard() error                      // ...this deletes it anyway (--force)

// Commit checks the staged content, freezes the stage with the version
// metadata, saves it, applies it, and deletes the file. If the object changed
// since the stage was created, nothing was written: the stage is unfrozen
// and saved as a draft.
func (f *File) Commit(ctx, root *ocfl.Root, msg string, user ocfl.User, opts ...CommitOption) error
func (f *File) Resume(ctx, root *ocfl.Root, opts ...CommitOption) error  // apply, delete
func (f *File) Revert(ctx, root *ocfl.Root, opts ...CommitOption) error  // revert, save as draft

func (s *Stage) Frozen() bool                             // replaces Finalized
func (s *Stage) freeze(msg string, user ocfl.User) error  // keeps a copy of the draft
func (s *Stage) unfreeze() error                          // restores that copy

func (c Cache) Create(id string, s *Stage) (*File, error) // MkdirAll, then Create
```

`CommitOption` carries the logger and the source-directory opener. When an
opener fails, `Content` wraps the error so it matches
`ocfl.ErrMissingContent`, and `errStagedContent` goes away.

`freeze`/`unfreeze` are unexported and only undo a freeze made in the same
process. Unfreezing without I/O is safe only when nothing has been written
to the object; the only case where we know that is a first apply that fails
with `ErrUpdateConflict`. A frozen stage read from disk may have been partly
applied, so the only way back to a draft is `Revert`. `freeze` keeps a copy
of the draft in memory using the existing `cloneUpdate`; `unfreeze` restores
it and errors for a stage opened already frozen. No ocfl-go change is
needed. (The alternative, `ObjectUpdate.Unfinalize()` in ocfl-go, would add a
public footgun and would have to be released first.)

## What `run` turns into

After a failed `Commit` or `Resume`, the stage's state picks the message;
`resuming`, `draft` and `discardConflict` go away:

```go
// commitError: s is the stage after the failed Commit or Resume.
switch {
case !s.Frozen():   // nothing was written (content check failed, or conflict on the first attempt)
case errors.Is(err, ocfl.ErrUpdateConflict):  // frozen: can't resume or revert -> delete --force
case errors.Is(err, context.Canceled):        // interrupted -> resume / revert
default:            // may be partially updated -> resume (after restoring files) / revert
}
```

`StageCommitCmd` becomes a switch over `f.Revert` / `f.Resume` / `f.Commit`
followed by `commitError`. `ocfl commit` becomes `Create`, then `Commit`;
if `Commit` fails and the stage isn't frozen, `Discard` it and tell the user
to run the command again.

## Smaller points

- `stageRef` shrinks to presentation only: `path`, `flags` and `shortID` for
  building command lines.
- Behavior change in `ocfl commit`: it now saves a draft before freezing. If
  the process crashes between `Create` and the freeze, a leftover draft
  blocks the next `ocfl commit` until it's deleted.
- Tests: CLI tests in `commit_test.go`/`stage_test.go` check stderr and
  should survive. `File.Commit`, `Resume` and `Revert` get unit tests in a
  new `stage/file_test.go`.
