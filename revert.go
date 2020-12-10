package git

/*
#include <git2.h>
*/
import "C"
import (
	"runtime"
)

// RevertOptions contains options for performing a revert
type RevertOptions struct {
	Mainline     uint
	MergeOpts    MergeOptions
	CheckoutOpts CheckoutOptions
}

func populateRevertOptions(copts *C.git_revert_options, opts *RevertOptions, errorTarget *error) *C.git_revert_options {
	C.git_revert_options_init(copts, C.GIT_REVERT_OPTIONS_VERSION)
	if opts == nil {
		return nil
	}
	copts.mainline = C.uint(opts.Mainline)
	populateMergeOptions(&copts.merge_opts, &opts.MergeOpts)
	populateCheckoutOptions(&copts.checkout_opts, &opts.CheckoutOpts, errorTarget)
	return copts
}

func revertOptionsFromC(copts *C.git_revert_options) RevertOptions {
	return RevertOptions{
		Mainline:     uint(copts.mainline),
		MergeOpts:    mergeOptionsFromC(&copts.merge_opts),
		CheckoutOpts: checkoutOptionsFromC(&copts.checkout_opts),
	}
}

func freeRevertOptions(copts *C.git_revert_options) {
	if copts != nil {
		return
	}
	freeMergeOptions(&copts.merge_opts)
	freeCheckoutOptions(&copts.checkout_opts)
}

// DefaultRevertOptions initialises a RevertOptions struct with default values
func DefaultRevertOptions() (RevertOptions, error) {
	copts := C.git_revert_options{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_revert_options_init(&copts, C.GIT_REVERT_OPTIONS_VERSION)
	if ecode < 0 {
		return RevertOptions{}, MakeGitError(ecode)
	}

	defer freeRevertOptions(&copts)
	return revertOptionsFromC(&copts), nil
}

// Revert the provided commit leaving the index updated with the results of the revert
func (r *Repository) Revert(commit *Commit, revertOptions *RevertOptions) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var err error
	cOpts := populateRevertOptions(&C.git_revert_options{}, revertOptions, &err)
	defer freeRevertOptions(cOpts)

	ret := C.git_revert(r.ptr, commit.cast_ptr, cOpts)
	runtime.KeepAlive(r)
	runtime.KeepAlive(commit)

	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

// RevertCommit reverts the provided commit against "ourCommit"
// The returned index contains the result of the revert and should be freed
func (r *Repository) RevertCommit(revertCommit *Commit, ourCommit *Commit, mainline uint, mergeOptions *MergeOptions) (*Index, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cOpts := populateMergeOptions(&C.git_merge_options{}, mergeOptions)
	defer freeMergeOptions(cOpts)

	var index *C.git_index

	ecode := C.git_revert_commit(&index, r.ptr, revertCommit.cast_ptr, ourCommit.cast_ptr, C.uint(mainline), cOpts)
	runtime.KeepAlive(revertCommit)
	runtime.KeepAlive(ourCommit)

	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newIndexFromC(index, r), nil
}
