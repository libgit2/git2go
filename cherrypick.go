package git

/*
#include <git2.h>
*/
import "C"
import (
	"runtime"
)

type CherrypickOptions struct {
	Version      uint
	Mainline     uint
	MergeOpts    MergeOptions
	CheckoutOpts CheckoutOptions
}

func cherrypickOptionsFromC(c *C.git_cherrypick_options) CherrypickOptions {
	opts := CherrypickOptions{
		Version:      uint(c.version),
		Mainline:     uint(c.mainline),
		MergeOpts:    mergeOptionsFromC(&c.merge_opts),
		CheckoutOpts: checkoutOptionsFromC(&c.checkout_opts),
	}
	return opts
}

func (opts *CherrypickOptions) toC() *C.git_cherrypick_options {
	if opts == nil {
		return nil
	}
	c := C.git_cherrypick_options{}
	c.version = C.uint(opts.Version)
	c.mainline = C.uint(opts.Mainline)
	c.merge_opts = *opts.MergeOpts.toC()
	c.checkout_opts = *opts.CheckoutOpts.toC()
	return &c
}

func freeCherrypickOpts(ptr *C.git_cherrypick_options) {
	if ptr == nil {
		return
	}
	freeCheckoutOptions(&ptr.checkout_opts)
}

func DefaultCherrypickOptions() (CherrypickOptions, error) {
	c := C.git_cherrypick_options{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_cherrypick_options_init(&c, C.GIT_CHERRYPICK_OPTIONS_VERSION)
	if ecode < 0 {
		return CherrypickOptions{}, MakeGitError(ecode)
	}
	defer freeCherrypickOpts(&c)
	return cherrypickOptionsFromC(&c), nil
}

func (v *Repository) Cherrypick(commit *Commit, opts CherrypickOptions) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cOpts := opts.toC()
	defer freeCherrypickOpts(cOpts)

	ecode := C.git_cherrypick(v.ptr, commit.cast_ptr, cOpts)
	runtime.KeepAlive(v)
	runtime.KeepAlive(commit)
	if ecode < 0 {
		return MakeGitError(ecode)
	}
	return nil
}

func (r *Repository) CherrypickCommit(pick, our *Commit, opts CherrypickOptions) (*Index, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cOpts := opts.MergeOpts.toC()

	var ptr *C.git_index
	ret := C.git_cherrypick_commit(&ptr, r.ptr, pick.cast_ptr, our.cast_ptr, C.uint(opts.Mainline), cOpts)
	runtime.KeepAlive(pick)
	runtime.KeepAlive(our)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return newIndexFromC(ptr, r), nil
}
