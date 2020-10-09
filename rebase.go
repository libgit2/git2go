package git

/*
#include <git2.h>

extern void _go_git_populate_commit_sign_cb(git_rebase_options *opts);
*/
import "C"
import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

// RebaseOperationType is the type of rebase operation
type RebaseOperationType uint

const (
	// RebaseOperationPick The given commit is to be cherry-picked.  The client should commit the changes and continue if there are no conflicts.
	RebaseOperationPick RebaseOperationType = C.GIT_REBASE_OPERATION_PICK
	// RebaseOperationReword The given commit is to be cherry-picked, but the client should prompt the user to provide an updated commit message.
	RebaseOperationReword RebaseOperationType = C.GIT_REBASE_OPERATION_REWORD
	// RebaseOperationEdit The given commit is to be cherry-picked, but the client should stop to allow the user to edit the changes before committing them.
	RebaseOperationEdit RebaseOperationType = C.GIT_REBASE_OPERATION_EDIT
	// RebaseOperationSquash The given commit is to be squashed into the previous commit.  The commit message will be merged with the previous message.
	RebaseOperationSquash RebaseOperationType = C.GIT_REBASE_OPERATION_SQUASH
	// RebaseOperationFixup No commit will be cherry-picked.  The client should run the given command and (if successful) continue.
	RebaseOperationFixup RebaseOperationType = C.GIT_REBASE_OPERATION_FIXUP
	// RebaseOperationExec No commit will be cherry-picked.  The client should run the given command and (if successful) continue.
	RebaseOperationExec RebaseOperationType = C.GIT_REBASE_OPERATION_EXEC
)

func (t RebaseOperationType) String() string {
	switch t {
	case RebaseOperationPick:
		return "pick"
	case RebaseOperationReword:
		return "reword"
	case RebaseOperationEdit:
		return "edit"
	case RebaseOperationSquash:
		return "squash"
	case RebaseOperationFixup:
		return "fixup"
	case RebaseOperationExec:
		return "exec"
	}
	return fmt.Sprintf("RebaseOperationType(%d)", t)
}

// Special value indicating that there is no currently active operation
var RebaseNoOperation uint = ^uint(0)

// Error returned if there is no current rebase operation
var ErrRebaseNoOperation = errors.New("no current rebase operation")

// RebaseOperation describes a single instruction/operation to be performed during the rebase.
type RebaseOperation struct {
	Type RebaseOperationType
	Id   *Oid
	Exec string
}

func newRebaseOperationFromC(c *C.git_rebase_operation) *RebaseOperation {
	operation := &RebaseOperation{}
	operation.Type = RebaseOperationType(c._type)
	operation.Id = newOidFromC(&c.id)
	operation.Exec = C.GoString(c.exec)

	return operation
}

//export commitSignCallback
func commitSignCallback(_signature *C.git_buf, _signature_field *C.git_buf, _commit_content *C.char, _payload unsafe.Pointer) C.int {
	opts, ok := pointerHandles.Get(_payload).(*RebaseOptions)
	if !ok {
		panic("invalid sign payload")
	}

	if opts.CommitSigningCallback == nil {
		return C.GIT_PASSTHROUGH
	}

	commitContent := C.GoString(_commit_content)

	signature, signatureField, err := opts.CommitSigningCallback(commitContent)
	if err != nil {
		if gitError, ok := err.(*GitError); ok {
			return C.int(gitError.Code)
		}
		return C.int(-1)
	}

	fillBuf := func(bufData string, buf *C.git_buf) error {
		clen := C.size_t(len(bufData))
		cstr := unsafe.Pointer(C.CString(bufData))
		defer C.free(cstr)

		// libgit2 requires the contents of the buffer to be NULL-terminated.
		// C.CString() guarantees that the returned buffer will be
		// NULL-terminated, so we can safely copy the terminator.
		if int(C.git_buf_set(buf, cstr, clen+1)) != 0 {
			return errors.New("could not set buffer")
		}

		return nil
	}

	if signatureField != "" {
		err := fillBuf(signatureField, _signature_field)
		if err != nil {
			return C.int(-1)
		}
	}

	err = fillBuf(signature, _signature)
	if err != nil {
		return C.int(-1)
	}

	return C.GIT_OK
}

// RebaseOptions are used to tell the rebase machinery how to operate
type RebaseOptions struct {
	Version               uint
	Quiet                 int
	InMemory              int
	RewriteNotesRef       string
	MergeOptions          MergeOptions
	CheckoutOptions       CheckoutOpts
	CommitSigningCallback CommitSigningCallback
}

// DefaultRebaseOptions returns a RebaseOptions with default values.
func DefaultRebaseOptions() (RebaseOptions, error) {
	opts := C.git_rebase_options{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_rebase_options_init(&opts, C.GIT_REBASE_OPTIONS_VERSION)
	if ecode < 0 {
		return RebaseOptions{}, MakeGitError(ecode)
	}
	return rebaseOptionsFromC(&opts), nil
}

func rebaseOptionsFromC(opts *C.git_rebase_options) RebaseOptions {
	return RebaseOptions{
		Version:         uint(opts.version),
		Quiet:           int(opts.quiet),
		InMemory:        int(opts.inmemory),
		RewriteNotesRef: C.GoString(opts.rewrite_notes_ref),
		MergeOptions:    mergeOptionsFromC(&opts.merge_options),
		CheckoutOptions: checkoutOptionsFromC(&opts.checkout_options),
	}
}

func (ro *RebaseOptions) toC() *C.git_rebase_options {
	if ro == nil {
		return nil
	}
	opts := &C.git_rebase_options{
		version:           C.uint(ro.Version),
		quiet:             C.int(ro.Quiet),
		inmemory:          C.int(ro.InMemory),
		rewrite_notes_ref: mapEmptyStringToNull(ro.RewriteNotesRef),
		merge_options:     *ro.MergeOptions.toC(),
		checkout_options:  *ro.CheckoutOptions.toC(),
	}

	if ro.CommitSigningCallback != nil {
		C._go_git_populate_commit_sign_cb(opts)
		opts.payload = pointerHandles.Track(ro)
	}

	return opts
}

func mapEmptyStringToNull(ref string) *C.char {
	if ref == "" {
		return nil
	}
	return C.CString(ref)
}

// Rebase is the struct representing a Rebase object.
type Rebase struct {
	ptr *C.git_rebase
	r   *Repository
}

// InitRebase initializes a rebase operation to rebase the changes in branch relative to upstream onto another branch.
func (r *Repository) InitRebase(branch *AnnotatedCommit, upstream *AnnotatedCommit, onto *AnnotatedCommit, opts *RebaseOptions) (*Rebase, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

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
	err := C.git_rebase_init(&ptr, r.ptr, branch.ptr, upstream.ptr, onto.ptr, opts.toC())
	runtime.KeepAlive(branch)
	runtime.KeepAlive(upstream)
	runtime.KeepAlive(onto)
	if err < 0 {
		return nil, MakeGitError(err)
	}

	return newRebaseFromC(ptr), nil
}

// OpenRebase opens an existing rebase that was previously started by either an invocation of InitRebase or by another client.
func (r *Repository) OpenRebase(opts *RebaseOptions) (*Rebase, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_rebase
	err := C.git_rebase_open(&ptr, r.ptr, opts.toC())
	runtime.KeepAlive(r)
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

// CurrentOperationIndex gets the index of the rebase operation that is
// currently being applied. There is also an error returned for API
// compatibility.
func (rebase *Rebase) CurrentOperationIndex() (uint, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var err error
	operationIndex := uint(C.git_rebase_operation_current(rebase.ptr))
	runtime.KeepAlive(rebase)
	if operationIndex == RebaseNoOperation {
		err = ErrRebaseNoOperation
	}

	return uint(operationIndex), err
}

// OperationCount gets the count of rebase operations that are to be applied.
func (rebase *Rebase) OperationCount() uint {
	ret := uint(C.git_rebase_operation_entrycount(rebase.ptr))
	runtime.KeepAlive(rebase)
	return ret
}

// Next performs the next rebase operation and returns the information about it.
// If the operation is one that applies a patch (which is any operation except RebaseOperationExec)
// then the patch will be applied and the index and working directory will be updated with the changes.
// If there are conflicts, you will need to address those before committing the changes.
func (rebase *Rebase) Next() (*RebaseOperation, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_rebase_operation
	err := C.git_rebase_next(&ptr, rebase.ptr)
	runtime.KeepAlive(rebase)
	if err < 0 {
		return nil, MakeGitError(err)
	}

	return newRebaseOperationFromC(ptr), nil
}

// Commit commits the current patch.
// You must have resolved any conflicts that were introduced during the patch application from the Next() invocation.
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
	defer C.git_signature_free(committerSig)

	cmsg := C.CString(message)
	defer C.free(unsafe.Pointer(cmsg))

	cerr := C.git_rebase_commit(ID.toC(), rebase.ptr, authorSig, committerSig, nil, cmsg)
	runtime.KeepAlive(ID)
	runtime.KeepAlive(rebase)
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
	runtime.KeepAlive(rebase)
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
	runtime.KeepAlive(rebase)
	if err < 0 {
		return MakeGitError(err)
	}
	return nil
}

// Free frees the Rebase object.
func (rebase *Rebase) Free() {
	runtime.SetFinalizer(rebase, nil)
	C.git_rebase_free(rebase.ptr)
}

func newRebaseFromC(ptr *C.git_rebase) *Rebase {
	rebase := &Rebase{ptr: ptr}
	runtime.SetFinalizer(rebase, (*Rebase).Free)
	return rebase
}
