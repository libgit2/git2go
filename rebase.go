package git

/*
#include <git2.h>

extern void _go_git_populate_rebase_callbacks(git_rebase_options *opts);
*/
import "C"
import (
	"errors"
	"fmt"
	"reflect"
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

//export commitCreateCallback
func commitCreateCallback(
	errorMessage **C.char,
	_out *C.git_oid,
	_author, _committer *C.git_signature,
	_message_encoding, _message *C.char,
	_tree *C.git_tree,
	_parent_count C.size_t,
	_parents **C.git_commit,
	handle unsafe.Pointer,
) C.int {
	data, ok := pointerHandles.Get(handle).(*rebaseOptionsData)
	if !ok {
		panic("invalid sign payload")
	}

	if data.options.CommitCreateCallback == nil && data.options.CommitSigningCallback == nil {
		return C.int(ErrorCodePassthrough)
	}

	messageEncoding := MessageEncodingUTF8
	if _message_encoding != nil {
		messageEncoding = MessageEncoding(C.GoString(_message_encoding))
	}
	tree := &Tree{
		Object: Object{
			ptr:  (*C.git_object)(_tree),
			repo: data.repo,
		},
		cast_ptr: _tree,
	}

	var goParents []*C.git_commit
	if _parent_count > 0 {
		hdr := reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(_parents)),
			Len:  int(_parent_count),
			Cap:  int(_parent_count),
		}
		goParents = *(*[]*C.git_commit)(unsafe.Pointer(&hdr))
	}

	parents := make([]*Commit, int(_parent_count))
	for i, p := range goParents {
		parents[i] = &Commit{
			Object: Object{
				ptr:  (*C.git_object)(p),
				repo: data.repo,
			},
			cast_ptr: p,
		}
	}

	if data.options.CommitCreateCallback != nil {
		oid, err := data.options.CommitCreateCallback(
			newSignatureFromC(_author),
			newSignatureFromC(_committer),
			messageEncoding,
			C.GoString(_message),
			tree,
			parents...,
		)
		if err != nil {
			if data.errorTarget != nil {
				*data.errorTarget = err
			}
			return setCallbackError(errorMessage, err)
		}
		if oid == nil {
			return C.int(ErrorCodePassthrough)
		}
		*_out = *oid.toC()
	} else if data.options.CommitSigningCallback != nil {
		commitContent, err := data.repo.CreateCommitBuffer(
			newSignatureFromC(_author),
			newSignatureFromC(_committer),
			messageEncoding,
			C.GoString(_message),
			tree,
			parents...,
		)
		if err != nil {
			if data.errorTarget != nil {
				*data.errorTarget = err
			}
			return setCallbackError(errorMessage, err)
		}

		signature, signatureField, err := data.options.CommitSigningCallback(string(commitContent))
		if err != nil {
			if data.errorTarget != nil {
				*data.errorTarget = err
			}
			return setCallbackError(errorMessage, err)
		}

		oid, err := data.repo.CreateCommitWithSignature(string(commitContent), signature, signatureField)
		if err != nil {
			if data.errorTarget != nil {
				*data.errorTarget = err
			}
			return setCallbackError(errorMessage, err)
		}
		*_out = *oid.toC()
	}

	return C.int(ErrorCodeOK)
}

// RebaseOptions are used to tell the rebase machinery how to operate.
type RebaseOptions struct {
	Quiet           int
	InMemory        int
	RewriteNotesRef string
	MergeOptions    MergeOptions
	CheckoutOptions CheckoutOptions
	// CommitCreateCallback is an optional callback that allows users to override
	// commit creation when rebasing. If specified, users can create
	// their own commit and provide the commit ID, which may be useful for
	// signing commits or otherwise customizing the commit creation. If this
	// callback returns a nil Oid, then the rebase will continue to create the
	// commit.
	CommitCreateCallback CommitCreateCallback
	// Deprecated: CommitSigningCallback is an optional callback that will be
	// called with the commit content, allowing a signature to be added to the
	// rebase commit. This field is only used when rebasing.  This callback is
	// not invoked if a CommitCreateCallback is specified.  CommitCreateCallback
	// should be used instead of this.
	CommitSigningCallback CommitSigningCallback
}

type rebaseOptionsData struct {
	options     *RebaseOptions
	repo        *Repository
	errorTarget *error
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
		Quiet:           int(opts.quiet),
		InMemory:        int(opts.inmemory),
		RewriteNotesRef: C.GoString(opts.rewrite_notes_ref),
		MergeOptions:    mergeOptionsFromC(&opts.merge_options),
		CheckoutOptions: checkoutOptionsFromC(&opts.checkout_options),
	}
}

func populateRebaseOptions(copts *C.git_rebase_options, opts *RebaseOptions, repo *Repository, errorTarget *error) *C.git_rebase_options {
	C.git_rebase_options_init(copts, C.GIT_REBASE_OPTIONS_VERSION)
	if opts == nil {
		return nil
	}

	copts.quiet = C.int(opts.Quiet)
	copts.inmemory = C.int(opts.InMemory)
	copts.rewrite_notes_ref = mapEmptyStringToNull(opts.RewriteNotesRef)
	populateMergeOptions(&copts.merge_options, &opts.MergeOptions)
	populateCheckoutOptions(&copts.checkout_options, &opts.CheckoutOptions, errorTarget)

	if opts.CommitCreateCallback != nil || opts.CommitSigningCallback != nil {
		data := &rebaseOptionsData{
			options:     opts,
			repo:        repo,
			errorTarget: errorTarget,
		}
		C._go_git_populate_rebase_callbacks(copts)
		copts.payload = pointerHandles.Track(data)
	}

	return copts
}

func freeRebaseOptions(copts *C.git_rebase_options) {
	if copts == nil {
		return
	}
	C.free(unsafe.Pointer(copts.rewrite_notes_ref))
	freeMergeOptions(&copts.merge_options)
	freeCheckoutOptions(&copts.checkout_options)
	if copts.payload != nil {
		pointerHandles.Untrack(copts.payload)
	}
}

func mapEmptyStringToNull(ref string) *C.char {
	if ref == "" {
		return nil
	}
	return C.CString(ref)
}

// Rebase is the struct representing a Rebase object.
type Rebase struct {
	doNotCompare
	ptr     *C.git_rebase
	r       *Repository
	options *C.git_rebase_options
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
	var err error
	cOpts := populateRebaseOptions(&C.git_rebase_options{}, opts, r, &err)
	ret := C.git_rebase_init(&ptr, r.ptr, branch.ptr, upstream.ptr, onto.ptr, cOpts)
	runtime.KeepAlive(branch)
	runtime.KeepAlive(upstream)
	runtime.KeepAlive(onto)
	if ret == C.int(ErrorCodeUser) && err != nil {
		freeRebaseOptions(cOpts)
		return nil, err
	}
	if ret < 0 {
		freeRebaseOptions(cOpts)
		return nil, MakeGitError(ret)
	}

	return newRebaseFromC(ptr, cOpts), nil
}

// OpenRebase opens an existing rebase that was previously started by either an invocation of InitRebase or by another client.
func (r *Repository) OpenRebase(opts *RebaseOptions) (*Rebase, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_rebase
	var err error
	cOpts := populateRebaseOptions(&C.git_rebase_options{}, opts, r, &err)
	ret := C.git_rebase_open(&ptr, r.ptr, cOpts)
	runtime.KeepAlive(r)
	if ret == C.int(ErrorCodeUser) && err != nil {
		freeRebaseOptions(cOpts)
		return nil, err
	}
	if ret < 0 {
		freeRebaseOptions(cOpts)
		return nil, MakeGitError(ret)
	}

	return newRebaseFromC(ptr, cOpts), nil
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
func (r *Rebase) Free() {
	runtime.SetFinalizer(r, nil)
	C.git_rebase_free(r.ptr)
	freeRebaseOptions(r.options)
}

func newRebaseFromC(ptr *C.git_rebase, opts *C.git_rebase_options) *Rebase {
	rebase := &Rebase{ptr: ptr, options: opts}
	runtime.SetFinalizer(rebase, (*Rebase).Free)
	return rebase
}
