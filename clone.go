package git

/*
#include <git2.h>

*/
import "C"
import (
	"runtime"
	"unsafe"
)

type CloneOptions struct {
	*CheckoutOpts
	*RemoteCallbacks
	Bare                 bool
	CheckoutBranch       string
	RemoteCreateCallback C.git_remote_create_cb
	RemoteCreatePayload  unsafe.Pointer
}

func Clone(url string, path string, options *CloneOptions) (*Repository, error) {
	repo := new(Repository)

	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	copts := (*C.git_clone_options)(C.calloc(1, C.size_t(unsafe.Sizeof(C.git_clone_options{}))))
	populateCloneOptions(copts, options)

	if len(options.CheckoutBranch) != 0 {
		copts.checkout_branch = C.CString(options.CheckoutBranch)
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ret := C.git_clone(&repo.ptr, curl, cpath, copts)
	freeCheckoutOpts(&copts.checkout_opts)
	C.free(unsafe.Pointer(copts.checkout_branch))
	C.free(unsafe.Pointer(copts))

	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(repo, (*Repository).Free)
	return repo, nil
}

func populateCloneOptions(ptr *C.git_clone_options, opts *CloneOptions) {
	C.git_clone_init_options(ptr, C.GIT_CLONE_OPTIONS_VERSION)

	if opts == nil {
		return
	}
	populateCheckoutOpts(&ptr.checkout_opts, opts.CheckoutOpts)
	populateRemoteCallbacks(&ptr.remote_callbacks, opts.RemoteCallbacks)
	ptr.bare = cbool(opts.Bare)

	if opts.RemoteCreateCallback != nil {
		ptr.remote_cb = opts.RemoteCreateCallback
		defer C.free(unsafe.Pointer(opts.RemoteCreateCallback))

		if opts.RemoteCreatePayload != nil {
			ptr.remote_cb_payload = opts.RemoteCreatePayload
			defer C.free(opts.RemoteCreatePayload)
		}
	}
}
