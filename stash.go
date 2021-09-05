package git

/*
#include <git2.h>

extern void _go_git_populate_stash_apply_callbacks(git_stash_apply_options *opts);
extern int _go_git_stash_foreach(git_repository *repo, void *payload);
*/
import "C"
import (
	"runtime"
	"unsafe"
)

// StashFlag are flags that affect the stash save operation.
type StashFlag int

const (
	// StashDefault represents no option, default.
	StashDefault StashFlag = C.GIT_STASH_DEFAULT

	// StashKeepIndex leaves all changes already added to the
	// index intact in the working directory.
	StashKeepIndex StashFlag = C.GIT_STASH_KEEP_INDEX

	// StashIncludeUntracked means all untracked files are also
	// stashed and then cleaned up from the working directory.
	StashIncludeUntracked StashFlag = C.GIT_STASH_INCLUDE_UNTRACKED

	// StashIncludeIgnored means all ignored files are also
	// stashed and then cleaned up from the working directory.
	StashIncludeIgnored StashFlag = C.GIT_STASH_INCLUDE_IGNORED
)

// StashCollection represents the possible operations that can be
// performed on the collection of stashes for a repository.
type StashCollection struct {
	doNotCompare
	repo *Repository
}

// Save saves the local modifications to a new stash.
//
// Stasher is the identity of the person performing the stashing.
// Message is the optional description along with the stashed state.
// Flags control the stashing process and are given as bitwise OR.
func (c *StashCollection) Save(
	stasher *Signature, message string, flags StashFlag) (*Oid, error) {

	oid := new(Oid)

	stasherC, err := stasher.toC()
	if err != nil {
		return nil, err
	}
	defer C.git_signature_free(stasherC)

	messageC := C.CString(message)
	defer C.free(unsafe.Pointer(messageC))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_stash_save(
		oid.toC(), c.repo.ptr,
		stasherC, messageC, C.uint32_t(flags))
	runtime.KeepAlive(c)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return oid, nil
}

// StashApplyFlag are flags that affect the stash apply operation.
type StashApplyFlag int

const (
	// StashApplyDefault is the default.
	StashApplyDefault StashApplyFlag = C.GIT_STASH_APPLY_DEFAULT

	// StashApplyReinstateIndex will try to reinstate not only the
	// working tree's changes, but also the index's changes.
	StashApplyReinstateIndex StashApplyFlag = C.GIT_STASH_APPLY_REINSTATE_INDEX
)

// StashApplyProgress are flags describing the progress of the apply operation.
type StashApplyProgress int

const (
	// StashApplyProgressNone means loading the stashed data from the object store.
	StashApplyProgressNone StashApplyProgress = C.GIT_STASH_APPLY_PROGRESS_NONE

	// StashApplyProgressLoadingStash means the stored index is being analyzed.
	StashApplyProgressLoadingStash StashApplyProgress = C.GIT_STASH_APPLY_PROGRESS_LOADING_STASH

	// StashApplyProgressAnalyzeIndex means the stored index is being analyzed.
	StashApplyProgressAnalyzeIndex StashApplyProgress = C.GIT_STASH_APPLY_PROGRESS_ANALYZE_INDEX

	// StashApplyProgressAnalyzeModified means the modified files are being analyzed.
	StashApplyProgressAnalyzeModified StashApplyProgress = C.GIT_STASH_APPLY_PROGRESS_ANALYZE_MODIFIED

	// StashApplyProgressAnalyzeUntracked means the untracked and ignored files are being analyzed.
	StashApplyProgressAnalyzeUntracked StashApplyProgress = C.GIT_STASH_APPLY_PROGRESS_ANALYZE_UNTRACKED

	// StashApplyProgressCheckoutUntracked means the untracked files are being written to disk.
	StashApplyProgressCheckoutUntracked StashApplyProgress = C.GIT_STASH_APPLY_PROGRESS_CHECKOUT_UNTRACKED

	// StashApplyProgressCheckoutModified means the modified files are being written to disk.
	StashApplyProgressCheckoutModified StashApplyProgress = C.GIT_STASH_APPLY_PROGRESS_CHECKOUT_MODIFIED

	// StashApplyProgressDone means the stash was applied successfully.
	StashApplyProgressDone StashApplyProgress = C.GIT_STASH_APPLY_PROGRESS_DONE
)

// StashApplyProgressCallback is the apply operation notification callback.
type StashApplyProgressCallback func(progress StashApplyProgress) error
type stashApplyProgressCallbackData struct {
	callback    StashApplyProgressCallback
	errorTarget *error
}

//export stashApplyProgressCallback
func stashApplyProgressCallback(progress C.git_stash_apply_progress_t, handle unsafe.Pointer) C.int {
	payload := pointerHandles.Get(handle)
	data, ok := payload.(*stashApplyProgressCallbackData)
	if !ok {
		panic("could not retrieve data for handle")
	}
	if data == nil || data.callback == nil {
		return C.int(ErrorCodeOK)
	}

	err := data.callback(StashApplyProgress(progress))
	if err != nil {
		*data.errorTarget = err
		return C.int(ErrorCodeUser)
	}
	return C.int(ErrorCodeOK)
}

// StashApplyOptions represents options to control the apply operation.
type StashApplyOptions struct {
	Flags            StashApplyFlag
	CheckoutOptions  CheckoutOptions            // options to use when writing files to the working directory
	ProgressCallback StashApplyProgressCallback // optional callback to notify the consumer of application progress
}

// DefaultStashApplyOptions initializes the structure with default values.
func DefaultStashApplyOptions() (StashApplyOptions, error) {
	optsC := C.git_stash_apply_options{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_stash_apply_options_init(&optsC, C.GIT_STASH_APPLY_OPTIONS_VERSION)
	if ecode < 0 {
		return StashApplyOptions{}, MakeGitError(ecode)
	}
	return StashApplyOptions{
		Flags:           StashApplyFlag(optsC.flags),
		CheckoutOptions: checkoutOptionsFromC(&optsC.checkout_options),
	}, nil
}

func populateStashApplyOptions(copts *C.git_stash_apply_options, opts *StashApplyOptions, errorTarget *error) *C.git_stash_apply_options {
	C.git_stash_apply_options_init(copts, C.GIT_STASH_APPLY_OPTIONS_VERSION)
	if opts == nil {
		return nil
	}
	copts.flags = C.uint32_t(opts.Flags)
	populateCheckoutOptions(&copts.checkout_options, &opts.CheckoutOptions, errorTarget)
	if opts.ProgressCallback != nil {
		progressData := &stashApplyProgressCallbackData{
			callback:    opts.ProgressCallback,
			errorTarget: errorTarget,
		}
		C._go_git_populate_stash_apply_callbacks(copts)
		copts.progress_payload = pointerHandles.Track(progressData)
	}
	return copts
}

func freeStashApplyOptions(copts *C.git_stash_apply_options) {
	if copts == nil {
		return
	}
	if copts.progress_payload != nil {
		pointerHandles.Untrack(copts.progress_payload)
	}
	freeCheckoutOptions(&copts.checkout_options)
}

// Apply applies a single stashed state from the stash list.
//
// If local changes in the working directory conflict with changes in the
// stash then ErrorCodeConflict will be returned.  In this case, the index
// will always remain unmodified and all files in the working directory will
// remain unmodified.  However, if you are restoring untracked files or
// ignored files and there is a conflict when applying the modified files,
// then those files will remain in the working directory.
//
// If passing the StashApplyReinstateIndex flag and there would be conflicts
// when reinstating the index, the function will return ErrorCodeConflict
// and both the working directory and index will be left unmodified.
//
// Note that a minimum checkout strategy of 'CheckoutSafe' is implied.
//
// 'index' is the position within the stash list. 0 points to the most
// recent stashed state.
//
// Returns error code ErrorCodeNotFound if there's no stashed state for the given
// index, error code ErrorCodeConflict if local changes in the working directory
// conflict with changes in the stash, the user returned error from the
// StashApplyProgressCallback, if any, or other error code.
//
// Error codes can be interogated with IsErrorCode(err, ErrorCodeNotFound).
func (c *StashCollection) Apply(index int, opts StashApplyOptions) error {
	var err error
	optsC := populateStashApplyOptions(&C.git_stash_apply_options{}, &opts, &err)
	defer freeStashApplyOptions(optsC)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_stash_apply(c.repo.ptr, C.size_t(index), optsC)
	runtime.KeepAlive(c)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

// StashCallback is called per entry when interating over all
// the stashed states.
//
// 'index' is the position of the current stash in the stash list,
// 'message' is the message used when creating the stash and 'id'
// is the commit id of the stash.
type StashCallback func(index int, message string, id *Oid) error
type stashCallbackData struct {
	callback    StashCallback
	errorTarget *error
}

//export stashForeachCallback
func stashForeachCallback(index C.size_t, message *C.char, id *C.git_oid, handle unsafe.Pointer) C.int {
	payload := pointerHandles.Get(handle)
	data, ok := payload.(*stashCallbackData)
	if !ok {
		panic("could not retrieve data for handle")
	}

	err := data.callback(int(index), C.GoString(message), newOidFromC(id))
	if err != nil {
		*data.errorTarget = err
		return C.int(ErrorCodeUser)
	}
	return C.int(ErrorCodeOK)
}

// Foreach loops over all the stashed states and calls the callback
// for each one.
//
// If callback returns an error, this will stop looping.
func (c *StashCollection) Foreach(callback StashCallback) error {
	var err error
	data := stashCallbackData{
		callback:    callback,
		errorTarget: &err,
	}
	handle := pointerHandles.Track(&data)
	defer pointerHandles.Untrack(handle)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C._go_git_stash_foreach(c.repo.ptr, handle)
	runtime.KeepAlive(c)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

// Drop removes a single stashed state from the stash list.
//
// 'index' is the position within the stash list. 0 points
// to the most recent stashed state.
//
// Returns error code ErrorCodeNotFound if there's no stashed
// state for the given index.
func (c *StashCollection) Drop(index int) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_stash_drop(c.repo.ptr, C.size_t(index))
	runtime.KeepAlive(c)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

// Pop applies a single stashed state from the stash list
// and removes it from the list if successful.
//
// 'index' is the position within the stash list. 0 points
// to the most recent stashed state.
//
// 'opts' controls how stashes are applied.
//
// Returns error code ErrorCodeNotFound if there's no stashed
// state for the given index.
func (c *StashCollection) Pop(index int, opts StashApplyOptions) error {
	var err error
	optsC := populateStashApplyOptions(&C.git_stash_apply_options{}, &opts, &err)
	defer freeStashApplyOptions(optsC)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_stash_pop(c.repo.ptr, C.size_t(index), optsC)
	runtime.KeepAlive(c)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}
