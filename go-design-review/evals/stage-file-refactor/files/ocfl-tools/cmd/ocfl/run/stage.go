package run

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/srerickson/ocfl-go"
	"github.com/srerickson/ocfl-go/digest"
	ocflfs "github.com/srerickson/ocfl-go/fs"

	"github.com/srerickson/ocfl-tools/cmd/ocfl/internal/diff"
	"github.com/srerickson/ocfl-tools/cmd/ocfl/internal/stage"
	"github.com/srerickson/ocfl-tools/cmd/ocfl/internal/statcache"
)

// StageCmd is 'ocfl stage': commands for preparing and committing object
// updates.
type StageCmd struct {
	stageCacheFlag `embed:""`

	Stage     string `name:"stage" short:"s" xor:"stage" help:"ID of a stage in the stage cache, or a unique prefix of it"`
	File      string `name:"file" short:"f" xor:"stage" help:"path to a stage file, used instead of the stage cache"`
	CheckStat string `name:"check-stat" enum:"default,minimal" default:"default" ocflenv:"OCFL_CHECK_STAT" help:"file metadata compared to find staged files that haven't changed: 'default' or 'minimal' (size, mtime in seconds, and mode; for network file systems)"`

	Add    StageAddCmd    `cmd:"" help:"Add a file or directory to the stage"`
	Commit StageCommitCmd `cmd:"" help:"Commit the stage as a new object version"`
	Delete StageDeleteCmd `cmd:"" help:"Delete a stage that hasn't been committed"`
	Diff   StageDiffCmd   `cmd:"" help:"Show changes between an upstream object or directory and the stage"`
	Ls     StageListCmd   `cmd:"" help:"List files in the stage state"`
	New    NewStageCmd    `cmd:"" help:"Create a new stage for preparing updates to an object"`
	Rm     StageRmCmd     `cmd:"" help:"Remove a file or directory from the stage"`
	Status StageStatusCmd `cmd:"" help:"Show stage details and report any errors"`
}

func (cmd *StageCmd) Help() string {
	return `A stage records changes to an object that can be committed as a new
version. Create one with 'stage new', change it with 'stage add' and 'stage
rm', and commit it with 'stage commit', which removes the stage. Staged
content is digested when it's added, but it isn't copied until the stage is
committed.

Stages are saved in the stage cache directory (--stage-cache), which has at
most one stage for each object in a storage root. 'stage new' prints the new
stage's ID, and the other commands take the ID, or a unique prefix of it,
with -s. To save a stage somewhere else, use -f with a file path instead.
Only 'stage new' uses --root: the other commands read the storage root from
the stage.

    ocfl stage new --id my-object
    ocfl stage add -s <stage-id> my-files/
    ocfl stage rm -s <stage-id> old-file.txt
    ocfl stage commit -s <stage-id> -m "update data" -n "My Name" -e me@example.net

The stage file records each staged file's metadata (size, modification
time, inode, and so on), like git's index. 'stage status' and 'stage add'
only read staged files whose metadata has changed. On network file systems
(SMB/CIFS, NFS), where the inode and change time aren't stable, use
--check-stat=minimal to compare only the size, modification time in whole
seconds, and mode.

If a commit is interrupted, the stage is kept. Run 'stage commit --resume'
to finish the commit, or 'stage commit --revert' to undo the partial update
and return the stage to its uncommitted state. 'stage delete' deletes a
stage that hasn't been committed.`
}

// ref returns the stage file named by --stage or --file. It returns an error
// if neither is set, or if --root was given on the command line: commands
// other than 'stage new' use the storage root saved in the stage.
func (cmd *StageCmd) ref(env *cmdEnv, kctx *kong.Context) (stageRef, error) {
	if flagGiven(kctx, "root") {
		return stageRef{}, errors.New("--root can only be used with 'stage new': the storage root is read from the stage")
	}
	switch {
	case cmd.File != "":
		return fileRef(cmd.File)
	case cmd.Stage != "":
		cache, err := cmd.cache(env)
		if err != nil {
			return stageRef{}, err
		}
		id, err := cache.Find(cmd.Stage)
		if err != nil {
			return stageRef{}, err
		}
		return cmd.cacheRef(kctx, cache, id)
	}
	return stageRef{}, errors.New("a stage is required: use --stage (-s) with a stage ID or --file (-f) with a stage file")
}

// load reads the stage file named by --stage or --file, and returns the stage
// and the stage file.
func (cmd *StageCmd) load(env *cmdEnv, kctx *kong.Context) (*stage.Stage, stageRef, error) {
	ref, err := cmd.ref(env, kctx)
	if err != nil {
		return nil, stageRef{}, err
	}
	s, err := cmd.read(ref)
	if err != nil {
		return nil, stageRef{}, err
	}
	return s, ref, nil
}

// loadDraft is like load, but it returns an error if the stage has been
// finalized for a commit, which must be resumed or reverted before the stage
// can be changed.
func (cmd *StageCmd) loadDraft(env *cmdEnv, kctx *kong.Context) (*stage.Stage, stageRef, error) {
	s, ref, err := cmd.load(env, kctx)
	if err != nil {
		return nil, stageRef{}, err
	}
	if s.Finalized() {
		return nil, stageRef{}, fmt.Errorf("the stage can't be changed: %w", inProgressError(s, ref))
	}
	return s, ref, nil
}

// read reads the stage file ref.
func (cmd *StageCmd) read(ref stageRef) (*stage.Stage, error) {
	s, err := stage.Read(ref.path)
	if err != nil {
		return nil, err
	}
	if cmd.CheckStat == "minimal" {
		s.Compare = statcache.CompareMinimal
	}
	return s, nil
}

// NewStageCmd is 'ocfl stage new'.
type NewStageCmd struct {
	Alg string `name:"alg" default:"sha512" help:"Digest Algorithm used to digest content. Ignored for existing objects."`
	ID  string `name:"id" short:"i" required:"" help:"object id for the new stage"`
}

func (cmd *NewStageCmd) Help() string {
	return `Creates a stage for the object in the storage root (--root) and prints its
stage ID. The stage is saved in the stage cache, or in the file given with
-f. There can only be one stage for an object in the stage cache: if one
exists, commit or delete it first.`
}

func (cmd *NewStageCmd) Run(ctx context.Context, env *cmdEnv, parent *StageCmd, kctx *kong.Context) error {
	if parent.Stage != "" {
		return errors.New("--stage (-s) can't be used with 'stage new': the stage ID is set from the storage root and object ID")
	}
	alg, err := digest.DefaultRegistry().Get(cmd.Alg)
	if err != nil {
		return err
	}
	root, rootLoc, err := env.getRootLocation(ctx)
	if err != nil {
		return err
	}
	var ref stageRef
	if parent.File != "" {
		ref, err = fileRef(parent.File)
	} else {
		ref, err = parent.objectRef(env, kctx, rootLoc, cmd.ID)
	}
	if err != nil {
		return err
	}
	// fail before reading the object if the stage exists
	if err := ref.checkNew(); err != nil {
		return err
	}
	s, err := newStage(ctx, root, rootLoc, cmd.ID, alg)
	if err != nil {
		return err
	}
	if err := ref.writeNew(s); err != nil {
		return err
	}
	if ref.shortID == "" {
		env.logger.Info("stage file created", "path", ref.path, "object_id", s.ID(), "object_version", s.NextHead())
		return nil
	}
	env.logger.Info("stage created", "stage", ref.shortID, "object_id", s.ID(), "object_version", s.NextHead(), "path", ref.path)
	fmt.Fprintln(env.stdout, ref.shortID)
	return nil
}

// StageAddCmd is 'ocfl stage add'.
type StageAddCmd struct {
	NoHidden bool   `name:"no-hidden" help:"exclude hidden files and directories (.*). Ignored if path is a file."`
	As       string `name:"as" help:"logical name for the new content. Default: base name if path is a file; '.' if path is a directory."`
	Jobs     int    `name:"jobs" short:"j" default:"0" help:"number of files to digest concurrently. Defaults to the number of CPU cores."`
	Remove   bool   `name:"remove" help:"also remove staged files not found in the path. Ignored if path is a file."`
	Path     string `arg:"" help:"file or parent directory for content to add to the stage"`
}

func (cmd *StageAddCmd) Run(ctx context.Context, env *cmdEnv, parent *StageCmd, kctx *kong.Context) error {
	s, ref, err := parent.loadDraft(env, kctx)
	if err != nil {
		return err
	}
	absPath, err := filepath.Abs(cmd.Path)
	if err != nil {
		return err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	ftype := info.Mode().Type()
	switch {
	case ftype.IsDir():
		opts := []stage.AddOption{
			stage.AddAs(cmd.As),
			stage.AddDigestJobs(cmd.Jobs),
		}
		if cmd.Remove {
			opts = append(opts, stage.AddAndRemove())
		}
		if cmd.NoHidden {
			opts = append(opts, stage.AddWithoutHidden())
		}
		if err := s.AddDir(ctx, absPath, env.logger, opts...); err != nil {
			return err
		}
	case ftype.IsRegular():
		if err := s.AddFile(ctx, absPath, env.logger, stage.AddAs(cmd.As)); err != nil {
			return err
		}
	default:
		return errors.New("unsupported file type for: " + absPath)
	}
	return stage.Write(ref.path, s)
}

// StageCommitCmd is 'ocfl stage commit'.
type StageCommitCmd struct {
	Message string `name:"message" short:"m" help:"Message to include in the object version metadata"`
	Name    string `name:"name" short:"n" ocflenv:"OCFL_USER_NAME" help:"Username to include in the object version metadata"`
	Email   string `name:"email" short:"e" ocflenv:"OCFL_USER_EMAIL" help:"User email to include in the object version metadata"`
	Resume  bool   `name:"resume" xor:"mode" help:"Finish an interrupted commit using the saved version metadata"`
	Revert  bool   `name:"revert" xor:"mode" help:"Undo an interrupted commit and return the stage to its uncommitted state"`
}

func (cmd *StageCommitCmd) Help() string {
	return `Commits the stage as a new object version. The version metadata (-m, -n
and -e) is saved in the stage file before the object is changed. If the
commit is interrupted or fails, the stage file is kept: run the command
with --resume to finish the commit using the saved values, or with --revert
to undo the partial update. -m, -n and -e can't be used with --resume or
--revert. After a revert, the stage can be changed and committed again, or
deleted. The commit is to the storage root saved in the stage.`
}

func (cmd *StageCommitCmd) Run(ctx context.Context, env *cmdEnv, parent *StageCmd, kctx *kong.Context) error {
	s, ref, err := parent.load(env, kctx)
	if err != nil {
		return err
	}
	root, err := env.openRoot(ctx, s.Root())
	if err != nil {
		return err
	}
	switch {
	case cmd.Revert:
		return cmd.revert(ctx, env, root, s, ref, kctx)
	case cmd.Resume:
		return cmd.resume(ctx, env, root, s, ref, kctx)
	case s.Finalized():
		return inProgressError(s, ref)
	}
	if err := checkWritable(root); err != nil {
		return err
	}
	user := newUser(cmd.Name, cmd.Email)
	// staged files are checked before the stage is finalized, so it can
	// still be changed if they aren't available.
	content, err := stagedContent(ctx, env, s)
	if err != nil {
		return fmt.Errorf("the stage can't be committed: %w\n  add the files again or remove them from the stage", err)
	}
	// the draft is saved again if the object has changed since the stage
	// was created, so the stage can be deleted.
	draft, err := parent.read(ref)
	if err != nil {
		return err
	}
	if err := s.Finalize(cmd.Message, user); err != nil {
		return err
	}
	// The finalized stage is saved before the object is changed: it's all
	// that's needed to resume or revert the commit.
	if err := stage.Write(ref.path, s); err != nil {
		return fmt.Errorf("saving the stage file before commit: %w", err)
	}
	return applyStage(ctx, env, root, s, ref, draft, false, content)
}

// resume finishes the interrupted commit of the finalized stage s, which is
// saved in the stage file ref, using the saved version metadata.
func (cmd *StageCommitCmd) resume(ctx context.Context, env *cmdEnv, root *ocfl.Root, s *stage.Stage, ref stageRef, kctx *kong.Context) error {
	if err := checkNoVersionFlags(kctx, "--resume"); err != nil {
		return err
	}
	if !s.Finalized() {
		return errors.New("stage isn't finalized: nothing to resume")
	}
	content, err := stagedContent(ctx, env, s)
	if err != nil {
		return commitError(root, s, ref, true, err)
	}
	return applyStage(ctx, env, root, s, ref, nil, true, content)
}

// revert reverts the interrupted commit of the finalized stage s, which is
// saved in the stage file ref, and saves the stage as a draft.
func (cmd *StageCommitCmd) revert(ctx context.Context, env *cmdEnv, root *ocfl.Root, s *stage.Stage, ref stageRef, kctx *kong.Context) error {
	if err := checkNoVersionFlags(kctx, "--revert"); err != nil {
		return err
	}
	if !s.Finalized() {
		return errors.New("stage isn't finalized: nothing to revert")
	}
	env.logger.Info("reverting object update", "object_id", s.ID(), "object_version", s.NextHead())
	err := root.Revert(ctx, s.Update(), ocfl.ApplyWithLogger(env.logger))
	switch {
	case errors.Is(err, ocfl.ErrUpdateCompleted):
		return fmt.Errorf("the commit of %s (%s) is complete and can't be reverted; to finish the commit and remove the stage file, run:\n  %s",
			s.ID(), s.NextHead(), ref.command("commit", "--resume"))
	case errors.Is(err, context.Canceled):
		return fmt.Errorf("revert of %s (%s) was interrupted\n  to finish reverting: %s",
			s.ID(), s.NextHead(), ref.command("commit", "--revert"))
	case errors.Is(err, ocfl.ErrUpdateConflict):
		return fmt.Errorf("the commit of %s (%s) can't be reverted: %w\n  object path: %s\n  check the object with 'ocfl validate', then delete the stage: %s",
			s.ID(), s.NextHead(), err, objectLocation(root, s.ID()), ref.command("delete", "--force"))
	case err != nil:
		return fmt.Errorf("reverting the commit of %s (%s): %w", s.ID(), s.NextHead(), err)
	}
	// s is a draft again
	if err := stage.Write(ref.path, s); err != nil {
		return fmt.Errorf("the commit was reverted, but the stage file couldn't be saved: %w", err)
	}
	env.logger.Info("object update reverted: the stage can be changed and committed again, or deleted",
		"path", ref.path, "delete", ref.command("delete"))
	return nil
}

// StageDeleteCmd is 'ocfl stage delete'.
type StageDeleteCmd struct {
	Force bool `name:"force" help:"Delete the stage even if a commit is in progress. The commit can't be resumed or reverted afterwards."`
}

func (cmd *StageDeleteCmd) Help() string {
	return `Deletes a stage that hasn't been committed. A stage with a commit in
progress can't be deleted: resume or revert the commit first. If neither is
possible (for example, because the object was changed by something else),
use --force. The object may be left partially updated: check it with 'ocfl
validate'.`
}

func (cmd *StageDeleteCmd) Run(env *cmdEnv, parent *StageCmd, kctx *kong.Context) error {
	s, ref, err := parent.load(env, kctx)
	if err != nil {
		return err
	}
	if s.Finalized() && !cmd.Force {
		return fmt.Errorf("the stage can't be deleted: %w", inProgressError(s, ref))
	}
	if err := os.Remove(ref.path); err != nil {
		return err
	}
	if s.Finalized() {
		env.logger.Warn("deleted a stage with a commit in progress: the object may be partially updated",
			"object_id", s.ID(), "object_version", s.NextHead(), "storage_root", s.Root())
		return nil
	}
	env.logger.Info("stage deleted", "object_id", s.ID(), "path", ref.path)
	return nil
}

// StageDiffCmd is 'ocfl stage diff'.
type StageDiffCmd struct {
	Dir string `name:"dir" help:"use a local directory rather than upstream object as basis for comparison to the stage."`
}

func (cmd *StageDiffCmd) Run(ctx context.Context, env *cmdEnv, parent *StageCmd, kctx *kong.Context) error {
	s, _, err := parent.load(env, kctx)
	if err != nil {
		return err
	}
	baseState := s.Update().BaseState().PathMap()
	if cmd.Dir != "" {
		baseState = ocfl.PathMap{}
		alg := s.DigestAlgorithm()
		filesIter, walkErr := ocflfs.UntilErr(ocflfs.WalkFiles(ctx, ocflfs.DirFS(cmd.Dir), "."))
		for result, err := range digest.DigestFiles(ctx, filesIter, alg) {
			if err != nil {
				return err
			}
			baseState[result.FullPath()] = result.Digests[alg.ID()]
		}
		if err := walkErr(); err != nil {
			return err
		}
	}
	diffs, err := diff.Diff(baseState, s.Update().NewState().PathMap())
	if err != nil {
		return err
	}
	if diffs.Empty() {
		return nil
	}
	fmt.Fprint(env.stdout, diffs.String())
	return nil
}

// StageListCmd is 'ocfl stage ls'.
type StageListCmd struct {
	WithDigests bool `name:"digests" short:"d" help:"include file digests in output"`
}

func (cmd *StageListCmd) Run(env *cmdEnv, parent *StageCmd, kctx *kong.Context) error {
	s, _, err := parent.load(env, kctx)
	if err != nil {
		return err
	}
	s.List(env.stdout, cmd.WithDigests)
	return nil
}

// StageRmCmd is 'ocfl stage rm'.
type StageRmCmd struct {
	Recursive bool   `name:"recursive" short:"r" help:"remove all files in the directory"`
	Path      string `arg:"" name:"path" help:"file or directory to remove"`
}

func (cmd *StageRmCmd) Run(env *cmdEnv, parent *StageCmd, kctx *kong.Context) error {
	s, ref, err := parent.loadDraft(env, kctx)
	if err != nil {
		return err
	}
	if err := s.Remove(cmd.Path, cmd.Recursive, env.logger); err != nil {
		return err
	}
	return stage.Write(ref.path, s)
}

// StageStatusCmd is 'ocfl stage status'.
type StageStatusCmd struct{}

func (cmd *StageStatusCmd) Help() string {
	return `Shows the object and version the stage is for, whether it has changes to
commit, and any staged files that have changed since they were added, like
'git status'. Staged files are only read if their metadata has changed
since they were added (see 'ocfl stage --help'). If a staged file has
changed and its content is needed for the commit, status exits with an
error.`
}

func (cmd *StageStatusCmd) Run(ctx context.Context, env *cmdEnv, parent *StageCmd, kctx *kong.Context) error {
	s, ref, err := parent.load(env, kctx)
	if err != nil {
		return err
	}
	upd := s.Update()
	fmt.Fprintf(env.stdout, "storage root: %s\n", s.Root())
	fmt.Fprintf(env.stdout, "object:       %s (%s)\n", s.ID(), s.NextHead())
	fmt.Fprintf(env.stdout, "digest alg:   %s\n", s.DigestAlgorithm().ID())
	fmt.Fprintf(env.stdout, "state size:   %d files\n", len(upd.NewState().PathMap()))
	stateDiff, err := diff.Diff(upd.BaseState().PathMap(), upd.NewState().PathMap())
	if err != nil {
		return err
	}
	switch {
	case s.Finalized():
		fmt.Fprintf(env.stdout, "commit in progress: %s\n", s.NextHead())
		fmt.Fprintf(env.stdout, "  to resume: %s\n", ref.command("commit", "--resume"))
		fmt.Fprintf(env.stdout, "  to revert: %s\n", ref.command("commit", "--revert"))
	case !stateDiff.Empty():
		fmt.Fprintln(env.stdout, "stage has changes to commit")
	default:
		fmt.Fprintln(env.stdout, "stage is unchanged and/or empty")
	}
	return checkStagedFiles(ctx, env, s, ref.path)
}

// stageCacheFlag is the --stage-cache flag of commands that save stages in the
// stage cache.
type stageCacheFlag struct {
	StageCache string `name:"stage-cache" ocflenv:"OCFL_STAGE_CACHE" help:"directory for saved stages. Default: 'ocfl-tools/stage' in the user's cache directory"`
}

// cache returns the stage cache: the --stage-cache directory, or by default,
// 'ocfl-tools/stage' in the user's cache directory.
func (f stageCacheFlag) cache(env *cmdEnv) (stage.Cache, error) {
	dir := f.StageCache
	if dir == "" {
		base, err := env.userCacheDir()
		if err != nil {
			return stage.Cache{}, fmt.Errorf("finding the default stage cache: %w; set --stage-cache or $OCFL_STAGE_CACHE", err)
		}
		dir = filepath.Join(base, "ocfl-tools", "stage")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return stage.Cache{}, err
	}
	return stage.Cache{Dir: abs}, nil
}

// objectRef returns the stage file in the stage cache for the object id in the
// storage root at the location rootLoc.
func (f stageCacheFlag) objectRef(env *cmdEnv, kctx *kong.Context, rootLoc, id string) (stageRef, error) {
	cache, err := f.cache(env)
	if err != nil {
		return stageRef{}, err
	}
	return f.cacheRef(kctx, cache, stage.ID(rootLoc, id))
}

// cacheRef returns the stage file for the stage ID id in cache. The stage is
// named by a short ID in commands, with --stage-cache if it was given on the
// command line.
func (f stageCacheFlag) cacheRef(kctx *kong.Context, cache stage.Cache, id string) (stageRef, error) {
	short, err := cache.ShortID(id)
	if err != nil {
		return stageRef{}, err
	}
	var flags []string
	if flagGiven(kctx, "stage-cache") {
		flags = append(flags, "--stage-cache", cache.Dir)
	}
	flags = append(flags, "-s", short)
	return stageRef{path: cache.Path(id), flags: flags, cache: cache, id: id, shortID: short}, nil
}

// existingStageError returns the error for creating a stage when the stage
// file ref already exists.
func existingStageError(ref stageRef) error {
	s, err := stage.Read(ref.path)
	if err != nil {
		return fmt.Errorf("stage file already exists, but it can't be read: %w\n  to start over, remove the stage file: %s", err, ref.path)
	}
	if s.Finalized() {
		return fmt.Errorf("a stage for %s already exists: %w", s.ID(), inProgressError(s, ref))
	}
	return fmt.Errorf("a stage for %s already exists: %s\n  to see it: %s\n  to delete it: %s",
		s.ID(), ref.path, ref.command("status"), ref.command("delete"))
}

// stageRef is a stage file and the flags that name it in 'ocfl stage'
// commands.
type stageRef struct {
	path  string
	flags []string

	// if the stage file is in the stage cache:
	cache   stage.Cache
	id      string // stage ID
	shortID string // the stage ID's shortest unique prefix, used in commands
}

// command returns the command line for the 'ocfl stage' subcommand sub, using
// r's stage file, with args.
func (r stageRef) command(sub string, args ...string) string {
	cmd := append([]string{"ocfl", "stage", sub}, r.flags...)
	cmd = append(cmd, args...)
	for i, arg := range cmd {
		cmd[i] = shellQuote(arg)
	}
	return strings.Join(cmd, " ")
}

// checkNew returns an error if r's stage file exists, so creating a stage can
// fail before doing any work.
func (r stageRef) checkNew() error {
	if _, err := os.Lstat(r.path); err == nil {
		return existingStageError(r)
	}
	return nil
}

// writeNew saves s in r's stage file, which must not exist: if it does, the
// error says how to see or delete the existing stage.
func (r stageRef) writeNew(s *stage.Stage) error {
	var err error
	if r.id != "" {
		err = r.cache.WriteNew(r.id, s)
	} else {
		err = stage.WriteNew(r.path, s)
	}
	if errors.Is(err, fs.ErrExist) {
		return existingStageError(r)
	}
	return err
}

// fileRef returns the stage file name, which is named with -f and its absolute
// path in commands.
func fileRef(name string) (stageRef, error) {
	abs, err := filepath.Abs(name)
	if err != nil {
		return stageRef{}, err
	}
	return stageRef{path: abs, flags: []string{"-f", abs}}, nil
}

// inProgressError returns an error saying that the commit of the finalized
// stage s, saved in the stage file ref, is in progress, and how to resume or
// revert it.
func inProgressError(s *stage.Stage, ref stageRef) error {
	return fmt.Errorf("a commit of %s (%s) is in progress\n  to resume: %s\n  to revert: %s",
		s.ID(), s.NextHead(), ref.command("commit", "--resume"), ref.command("commit", "--revert"))
}

// checkNoVersionFlags returns an error if -m, -n or -e were given on the
// command line. mode is the flag (--resume or --revert) that can't be used
// with them.
func checkNoVersionFlags(kctx *kong.Context, mode string) error {
	// -n and -e may be set from the environment: only flags on the command
	// line are an error.
	for _, flag := range []string{"message", "name", "email"} {
		if flagGiven(kctx, flag) {
			return fmt.Errorf("%s can't be used with --message (-m), --name (-n) or --email (-e)", mode)
		}
	}
	return nil
}

// checkStagedFiles reports staged files in s that have changed since they were
// added, and saves the refreshed stats of unchanged files in the stage file
// stagePath. It returns an error if the commit needs content that is only in
// changed files.
func checkStagedFiles(ctx context.Context, env *cmdEnv, s *stage.Stage, stagePath string) error {
	result, err := s.Check(ctx)
	if err != nil {
		return err
	}
	if result.Refreshed {
		// like 'git status', saving the refreshed stats is opportunistic
		if err := stage.Write(stagePath, s); err != nil {
			env.logger.Warn("couldn't save refreshed file stats in the stage file", "err", err)
		}
	}
	if len(result.Changed) > 0 {
		fmt.Fprintln(env.stdout, "staged files changed since they were added:")
		if !s.Finalized() {
			fmt.Fprintln(env.stdout, `  (use "ocfl stage add" to stage the new content)`)
		}
		for _, c := range result.Changed {
			if c.Err != nil {
				fmt.Fprintf(env.stdout, "  error:    %s: %v\n", c.Path, c.Err)
				continue
			}
			fmt.Fprintf(env.stdout, "  %-9s %s (%s)\n", c.Status.String()+":", c.Path, c.Source)
		}
	}
	if len(result.Unavailable) > 0 {
		todo := "add the files again or remove them from the stage"
		if s.Finalized() {
			todo = "restore the files to resume the commit, or revert it"
		}
		return fmt.Errorf("the stage can't be committed: %w\n  %s", &stage.UnavailableError{Paths: result.Unavailable}, todo)
	}
	return nil
}
