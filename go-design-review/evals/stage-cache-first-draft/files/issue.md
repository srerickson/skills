# stage, commit: persistent stage cache and stage IDs (-s / --stage)

Store stage files in a persistent stage cache directory by default so both `ocfl commit` and `ocfl stage` share a unified, reboot-safe mechanism for managing draft and interrupted stages (replaces the temp-directory stage file used by `ocfl commit`).

## Design

### Stage cache directory & stage IDs

* Add a `--stage-cache` flag (with `$OCFL_STAGE_CACHE` env var) that sets the directory where cached stage files are saved.
* Default to `<os.UserCacheDir()>/ocfl-tools/stage` across macOS, Linux, and Windows.
* Compute each cached stage's ID deterministically from the storage root and object ID:
  * Digest input: `"root:" + rootLoc + "\nid:" + id`, where `rootLoc` is the canonical root location string.
  * Encode the SHA-256 digest using lowercase unpadded Base32 (52 characters), stored as `<stage-id>.json` in the stage cache directory.
* Check for an existing stage file before overwriting: if a stage (draft or finalized) already exists for `(root, id)`, refuse to overwrite it and surface recovery/management commands.

### Storing root location in stage files

* Record the canonical storage root location (`root`) in the stage JSON alongside `update`, `sources`, and `files`.
* `--root` is only valid for `stage new` (and `ocfl commit`). Other `stage` subcommands read the storage root from the stage file and return an error if `--root` was explicitly passed.

### `ocfl stage` flags and subcommands

* `ocfl stage` takes `-s` / `--stage` (stage ID prefix in `--stage-cache`) or `-f` / `--file` (explicit path to a stage file). They are mutually exclusive.
* `stage new` saves to the stage cache by default (or to `-f`), prints the stage ID prefix; `-s` is an error.
* All other `stage` subcommands (`add`, `rm`, `status`, `ls`, `diff`, `commit`, `delete`) require `-s` or `-f`.
* New subcommand `stage delete`: deletes a draft stage; refuses to delete a finalized stage (which has an in-progress commit that must be resumed or reverted first) and prints the resume/revert commands.

### Recovery commands

* When `ocfl commit` or `ocfl stage commit` is interrupted or fails, keep the stage file and print `ocfl stage commit -s <prefix>` (or `-f <path>`) commands to resume or revert.
