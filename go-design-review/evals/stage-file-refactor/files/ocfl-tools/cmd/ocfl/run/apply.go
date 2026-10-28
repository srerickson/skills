package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/srerickson/ocfl-go"
	"github.com/srerickson/ocfl-go/digest"
	ocflfs "github.com/srerickson/ocfl-go/fs"

	"github.com/srerickson/ocfl-tools/cmd/ocfl/internal/stage"
)

// errStagedContent wraps errors opening the directories staged files were
// added from.
var errStagedContent = errors.New("staged content isn't available")

// checkWritable returns an error if root's storage can't be written, so a
// commit fails before its stage is finalized.
func checkWritable(root *ocfl.Root) error {
	if _, ok := root.FS().(ocflfs.WriteFS); !ok {
		return fmt.Errorf("storage root %s is read-only", locationString(root.FS(), root.Path()))
	}
	return nil
}

func newUser(name string, email string) ocfl.User {
	if email != "" && !strings.HasPrefix(email, "email:") {
		email = "email:" + email
	}
	return ocfl.User{Name: name, Address: email}
}

// newStage starts a stage to the object id in root, which is at the location
// rootLoc. alg is used if the object doesn't exist.
func newStage(ctx context.Context, root *ocfl.Root, rootLoc string, id string, alg digest.Algorithm) (*stage.Stage, error) {
	s, err := stage.New(ctx, root, rootLoc, id, alg)
	if errors.Is(err, ocfl.ErrObjectIncomplete) {
		return nil, fmt.Errorf("%w\nif a commit to the object was interrupted, resume or revert it using its stage file", err)
	}
	return s, err
}

// stagedContent returns the content source for committing s: its unchanged
// staged files.
func stagedContent(ctx context.Context, env *cmdEnv, s *stage.Stage) (ocfl.ContentSource, error) {
	return s.Content(ctx, func(dir string) (ocflfs.FS, error) {
		fsys, err := env.localFS(dir)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errStagedContent, err)
		}
		return fsys, nil
	})
}

// applyStage applies the finalized update in s, which is saved in the stage
// file ref, with content from staged files, and removes the stage file.
// resuming is true if an earlier attempt to apply s may have changed the
// object. If the update isn't applied, the stage file is kept and the error
// says how to resume or revert the commit.
//
// If the first attempt (resuming is false) finds that the object has changed
// since the stage was created, nothing was written, and the stage can never
// be committed. The stage file is replaced with draft, the stage before it was
// finalized, or removed if draft is nil.
func applyStage(ctx context.Context, env *cmdEnv, root *ocfl.Root, s *stage.Stage, ref stageRef, draft *stage.Stage, resuming bool, content ocfl.ContentSource) error {
	env.logger.Info("starting object update", "object_id", s.ID(), "object_version", s.NextHead())
	if _, err := root.Apply(ctx, s.Update(), content, ocfl.ApplyWithLogger(env.logger)); err != nil {
		if !resuming && errors.Is(err, ocfl.ErrUpdateConflict) {
			return discardConflict(s, ref, draft, err)
		}
		return commitError(root, s, ref, resuming, err)
	}
	env.logger.Info("object update complete", "object_id", s.ID(), "object_version", s.NextHead())
	if err := os.Remove(ref.path); err != nil {
		return fmt.Errorf("removing stage file: %w", err)
	}
	return nil
}

// commitError describes the error err from applying the stage s, saved in the
// stage file ref, and says what to do next. resuming is true if an earlier
// attempt to apply s may have changed the object.
func commitError(root *ocfl.Root, s *stage.Stage, ref stageRef, resuming bool, err error) error {
	commit := fmt.Sprintf("commit of %s (%s)", s.ID(), s.NextHead())
	objPath := "  object path: " + objectLocation(root, s.ID())
	resume := ref.command("commit", "--resume")
	revert := "  to revert: " + ref.command("commit", "--revert")
	var lines []string
	switch {
	case errors.Is(err, context.Canceled):
		lines = []string{commit + " was interrupted; the object may be partially updated",
			objPath, "  to resume: " + resume, revert}
	case errors.Is(err, ocfl.ErrUpdateConflict):
		// the object was changed by something else after an interrupted
		// attempt: the stage can't be resumed or reverted.
		lines = []string{commit + " can't be resumed or reverted: " + err.Error(),
			objPath, "  check the object with 'ocfl validate', then delete the stage: " + ref.command("delete", "--force")}
	case errors.Is(err, ocfl.ErrContentChanged),
		resuming && (errors.Is(err, ocfl.ErrMissingContent) || errors.Is(err, errStagedContent)):
		// Apply finds changed content as it copies it, after other content
		// may have been written, and a resumed commit may have been
		// interrupted after changing the object.
		lines = []string{commit + " failed; the object may be partially updated: " + err.Error(),
			objPath, "  to resume after restoring the staged files: " + resume, revert}
	case errors.Is(err, ocfl.ErrMissingContent), errors.Is(err, errStagedContent):
		// Apply checks for missing content before it writes anything.
		lines = []string{commit + " failed before the object was changed: " + err.Error(),
			"  to resume after restoring the staged files: " + resume, revert}
	default:
		lines = []string{commit + " failed; the object may be partially updated: " + err.Error(),
			objPath, "  to resume: " + resume, revert}
	}
	return &wrappedError{msg: strings.Join(lines, "\n"), err: err}
}

// objectLocation returns the storage location of the object id in root, for
// messages.
func objectLocation(root *ocfl.Root, id string) string {
	objPath, err := root.ResolveID(id)
	if err != nil {
		return "unknown: " + err.Error()
	}
	return locationString(root.FS(), path.Join(root.Path(), objPath))
}

// discardConflict replaces the finalized stage s, saved in the stage file ref,
// with draft, or removes it if draft is nil, after the error err from the
// first attempt to apply s: the object has changed since the stage was
// created.
func discardConflict(s *stage.Stage, ref stageRef, draft *stage.Stage, err error) error {
	lines := []string{fmt.Sprintf("commit of %s (%s) failed before the object was changed: %v", s.ID(), s.NextHead(), err)}
	if draft == nil {
		if rmErr := os.Remove(ref.path); rmErr != nil {
			lines = append(lines, "  the stage file couldn't be removed: "+rmErr.Error(),
				"  to delete it: "+ref.command("delete", "--force"))
		} else {
			lines = append(lines, "  to commit to the object's current version, run the command again")
		}
	} else if wErr := stage.Write(ref.path, draft); wErr != nil {
		lines = append(lines, "  the stage file couldn't be restored: "+wErr.Error(),
			"  to delete it: "+ref.command("delete", "--force"))
	} else {
		lines = append(lines, "  the stage can't be committed: delete it and create a new one",
			"  to delete it: "+ref.command("delete"))
	}
	return &wrappedError{msg: strings.Join(lines, "\n"), err: err}
}
