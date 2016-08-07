package git

/*
#include <git2.h>
*/
import "C"
import (
	"errors"
	"runtime"
	"unsafe"
)

// RebaseOperationType is the type of rebase operation
type RebaseOperationType uint

const (
	// RebaseOperationPick The given commit is to be cherry-picked.  The client should commit the changes and continue if there are no conflicts.
	RebaseOperationPick RebaseOperationType = C.GIT_REBASE_OPERATION_PICK
	// RebaseOperationEdit The given commit is to be cherry-picked, but the client should stop to allow the user to edit the changes before committing them.
	RebaseOperationEdit RebaseOperationType = C.GIT_REBASE_OPERATION_EDIT
	// RebaseOperationSquash The given commit is to be squashed into the previous commit.  The commit message will be merged with the previous message.
	RebaseOperationSquash RebaseOperationType = C.GIT_REBASE_OPERATION_SQUASH
	// RebaseOperationFixup No commit will be cherry-picked.  The client should run the given command and (if successful) continue.
	RebaseOperationFixup RebaseOperationType = C.GIT_REBASE_OPERATION_FIXUP
	// RebaseOperationExec No commit will be cherry-picked.  The client should run the given command and (if successful) continue.
	RebaseOperationExec RebaseOperationType = C.GIT_REBASE_OPERATION_EXEC
)

// RebaseOperation describes a single instruction/operation to be performed during the rebase.
type RebaseOperation struct {
	Type RebaseOperationType
	ID   *Oid
	Exec string
}

func newRebaseOperationFromC(c *C.git_rebase_operation) *RebaseOperation {
	operation := &RebaseOperation{}
	operation.Type = RebaseOperationType(c._type)
	operation.ID = newOidFromC(&c.id)
	operation.Exec = C.GoString(c.exec)

	return operation
}

// RebaseOptions are used to tell the rebase machinery how to operate
type RebaseOptions struct{}

// Rebase object wrapper for C pointer
type Rebase struct {
	ptr *C.git_rebase
}

//RebaseInit initializes a rebase operation to rebase the changes in branch relative to upstream onto another branch.
func (r *Repository) RebaseInit(branch *AnnotatedCommit, upstream *AnnotatedCommit, onto *AnnotatedCommit, opts *RebaseOptions) (*Rebase, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	//TODO : use real rebase_options
	if opts != nil {
		return nil, errors.New("RebaseOptions Not implemented yet, use nil for default opts")
	}

	if branch == nil {
		branch = &AnnotatedCommit{ptr: nil}
	}

	if upstream == nil {
		upstream = &AnnotatedCommit{ptr: nil}
	}

	if onto == nil {
		onto = &AnnotatedCommit{ptr: nil}
	}

	var ptr *C.git_rebase
	err := C.git_rebase_init(&ptr, r.ptr, branch.ptr, upstream.ptr, onto.ptr, nil)
	if err < 0 {
		return nil, MakeGitError(err)
	}

	return newRebaseFromC(ptr), nil
}

// OperationAt gets the rebase operation specified by the given index.
func (rebase *Rebase) OperationAt(index uint) *RebaseOperation {
	operation := C.git_rebase_operation_byindex(rebase.ptr, C.size_t(index))
	return newRebaseOperationFromC(operation)
}

// CurrentOperationIndex gets the index of the rebase operation that is currently being applied.
// If the first operation has not yet been applied then this returns -1 (C.GIT_REBASE_NO_OPERATION).
func (rebase *Rebase) CurrentOperationIndex() int {
	return int(C.git_rebase_operation_current(rebase.ptr))
}

// OperationCount gets the count of rebase operations that are to be applied.
func (rebase *Rebase) OperationCount() uint {
	return uint(C.git_rebase_operation_entrycount(rebase.ptr))
}

// Next performs the next rebase operation and returns the information about it.
// If the operation is one that applies a patch (which is any operation except GIT_REBASE_OPERATION_EXEC)
// then the patch will be applied and the index and working directory will be updated with the changes.
// If there are conflicts, you will need to address those before committing the changes.
func (rebase *Rebase) Next() (*RebaseOperation, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_rebase_operation
	err := C.git_rebase_next(&ptr, rebase.ptr)
	if err < 0 {
		return nil, MakeGitError(err)
	}

	return newRebaseOperationFromC(ptr), nil
}

// Commit commits the current patch.
// You must have resolved any conflicts that were introduced during the patch application from the git_rebase_next invocation.
func (rebase *Rebase) Commit(ID *Oid, author, committer *Signature, message string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	authorSig, err := author.toC()
	if err != nil {
		return err
	}
	defer C.git_signature_free(authorSig)

	committerSig, err := committer.toC()
	if err != nil {
		return err
	}

	cmsg := C.CString(message)
	defer C.free(unsafe.Pointer(cmsg))

	cerr := C.git_rebase_commit(ID.toC(), rebase.ptr, authorSig, committerSig, nil, cmsg)
	if cerr < 0 {
		return MakeGitError(cerr)
	}

	return nil
}

// Finish finishes a rebase that is currently in progress once all patches have been applied.
func (rebase *Rebase) Finish() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C.git_rebase_finish(rebase.ptr, nil)
	if err < 0 {
		return MakeGitError(err)
	}

	return nil
}

// Abort aborts a rebase that is currently in progress, resetting the repository and working directory to their state before rebase began.
func (rebase *Rebase) Abort() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C.git_rebase_abort(rebase.ptr)
	if err < 0 {
		return MakeGitError(err)
	}
	return nil
}

//Free frees the Rebase object and underlying git_rebase C pointer.
func (rebase *Rebase) Free() {
	runtime.SetFinalizer(rebase, nil)
	C.git_reference_free(rebase.ptr)
}

func newRebaseFromC(ptr *C.git_rebase) *Rebase {
	rebase := &Rebase{ptr: ptr}
	runtime.SetFinalizer(rebase, (*Rebase).Free)
	return rebase
}

/* TODO -- Add last wrapper services and manage rebase_options

int git_rebase_init_options(git_rebase_options *opts, unsigned int version);
int git_rebase_open(git_rebase **out, git_repository *repo, const git_rebase_options *opts);

*/
