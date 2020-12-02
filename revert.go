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

func (opts *RevertOptions) toC(errorTarget *error) *C.git_revert_options {
	if opts == nil {
		return nil
	}
	return &C.git_revert_options{
		version:       C.GIT_REVERT_OPTIONS_VERSION,
		mainline:      C.uint(opts.Mainline),
		merge_opts:    *opts.MergeOpts.toC(),
		checkout_opts: *opts.CheckoutOpts.toC(errorTarget),
	}
}

func revertOptionsFromC(opts *C.git_revert_options) RevertOptions {
	return RevertOptions{
		Mainline:     uint(opts.mainline),
		MergeOpts:    mergeOptionsFromC(&opts.merge_opts),
		CheckoutOpts: checkoutOptionsFromC(&opts.checkout_opts),
	}
}

func freeRevertOptions(opts *C.git_revert_options) {
	if opts != nil {
		return
	}
	freeMergeOptions(&opts.merge_opts)
	freeCheckoutOptions(&opts.checkout_opts)
}

// DefaultRevertOptions initialises a RevertOptions struct with default values
func DefaultRevertOptions() (RevertOptions, error) {
	opts := C.git_revert_options{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_revert_options_init(&opts, C.GIT_REVERT_OPTIONS_VERSION)
	if ecode < 0 {
		return RevertOptions{}, MakeGitError(ecode)
	}

	defer freeRevertOptions(&opts)
	return revertOptionsFromC(&opts), nil
}

// Revert the provided commit leaving the index updated with the results of the revert
func (r *Repository) Revert(commit *Commit, revertOptions *RevertOptions) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var err error
	cOpts := revertOptions.toC(&err)
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

	cOpts := mergeOptions.toC()
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
