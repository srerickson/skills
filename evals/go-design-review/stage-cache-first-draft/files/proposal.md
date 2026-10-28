# Proposal: persistent stage cache and stage IDs

Implements the issue in `issue.md`. The current code is in `ocfl-tools/`
(ocfl-go v0.13.0 is in the Go module cache).

## stage package (`cmd/ocfl/internal/stage`)

New file `cache.go`:

```go
// ID returns the stage ID for the object id in the storage root at rootLoc:
// lowercase unpadded base32 SHA-256 of "root:" + rootLoc + "\nid:" + id.
func ID(rootLoc, id string) string

// Cache is a directory of stage files named <stage-id>.json.
type Cache struct{ Dir string }

func (c Cache) Path(id string) string
// WriteNew creates the cache directory if needed and saves s as the stage
// file for id. It fails with fs.ErrExist if the file exists.
func (c Cache) WriteNew(id string, s *Stage) error
// Find returns the stage ID with the given prefix; errors if none or
// ambiguous. Temp files and other files are ignored.
func (c Cache) Find(prefix string) (string, error)
// ShortID returns the shortest prefix of id (at least 7 chars) that is
// unique in the cache.
func (c Cache) ShortID(id string) (string, error)
```

Changes to `stage.go`:

```go
// New takes the canonical root location, which is saved in the stage.
func New(ctx context.Context, root *ocfl.Root, rootLoc string, id string, alg digest.Algorithm) (*Stage, error)
func (s *Stage) Root() string // the storage root location

// Read and Write stay as they are. WriteNew is like Write but doesn't
// replace an existing file (hard link, falling back to rename).
func Read(name string) (*Stage, error)
func Write(name string, s *Stage) error
func WriteNew(name string, s *Stage) error
```

The stage JSON gets a required `root` field. `Finalize(msg, user)` and
`Finalized()` are unchanged.

## run package (`cmd/ocfl/run`)

`StageCmd` gets `-s/--stage` and `-f/--file` (mutually exclusive) and an
embedded `stageCacheFlag` (`--stage-cache`, `$OCFL_STAGE_CACHE`, default
`<user cache dir>/ocfl-tools/stage`). `CommitCmd` embeds `stageCacheFlag`
too.

```go
// stageRef is a resolved stage file and the flags that name it in printed
// commands.
type stageRef struct {
	path  string
	id    string   // short ID, if from the cache
	flags []string // e.g. ["--stage-cache", dir, "-s", short] or ["-f", path]
}
func (r stageRef) command(sub string, args ...string) string

// ref resolves -s or -f; errors if --root was given on the command line.
func (cmd *StageCmd) ref(env *cmdEnv, kctx *kong.Context) (stageRef, error)
func (cmd *StageCmd) read(ref stageRef) (*stage.Stage, error)
// readDraft is read, but fails if the stage is finalized.
func (cmd *StageCmd) readDraft(ref stageRef) (*stage.Stage, error)

func (f stageCacheFlag) cache(env *cmdEnv) (stage.Cache, error)
func (f stageCacheFlag) cacheRef(kctx *kong.Context, c stage.Cache, id string) (stageRef, error)

func inProgressError(s *stage.Stage, ref stageRef) error // resume/revert commands
func existingStageError(ref stageRef) error             // status/delete or resume/revert commands
```

Shared by `ocfl commit` and `ocfl stage commit`:

```go
// stagedContent opens the directories files were staged from.
func stagedContent(ctx context.Context, env *cmdEnv, s *stage.Stage) (ocfl.ContentSource, error)

// applyStage applies the finalized s, saved in the stage file ref, then
// removes the stage file. resuming is true if an earlier attempt may have
// changed the object. If the update isn't applied, the stage file is kept
// and the error says how to resume or revert.
func applyStage(ctx context.Context, env *cmdEnv, root *ocfl.Root, s *stage.Stage, ref stageRef, resuming bool, content ocfl.ContentSource) error

// commitError builds the message for a failed apply: interrupted, update
// conflict ("the stage file was kept"), missing or changed content, or
// other, with resume/revert commands from ref.
func commitError(root *ocfl.Root, s *stage.Stage, ref stageRef, resuming bool, err error) error
```

### Flows

- `stage new`: resolve the cache ref for `stage.ID(rootLoc, id)`; if the
  file exists, `existingStageError`; `stage.New`; `cache.WriteNew`; print
  the short ID.
- `stage add/rm`: `ref`, `readDraft`, change the stage, `stage.Write`.
- `stage status`: `ref`, `read`, check stats, `stage.Write` to save
  refreshed stats.
- `stage commit`: `ref`, `read`, open root from `s.Root()`. With
  `--resume`: `stagedContent`, `applyStage(..., resuming=true, ...)`. With
  `--revert`: `root.Revert`, then `stage.Write` the reverted (draft) stage.
  Otherwise: refuse if finalized; `stagedContent`; `s.Finalize`;
  `stage.Write`; `applyStage(..., resuming=false, ...)`.
- `stage delete`: `ref`, `read`; refuse if finalized (print resume/revert
  commands); `os.Remove(ref.path)`.
- `ocfl commit`: compute the cache ref; refuse if it exists; `stage.New`;
  `AddDir`; `stagedContent`; `s.Finalize`; `cache.WriteNew`;
  `applyStage(..., resuming=false, ...)`.
