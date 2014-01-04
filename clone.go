package git

/*
#include <git2.h>
#include <git2/errors.h>

static git_clone_options git_clone_options_init() {
	git_clone_options ret = GIT_CLONE_OPTIONS_INIT;
	return ret;
}

*/
import "C"
import (
	"runtime"
	"unsafe"
)

type CloneOptions struct {
	*CheckoutOpts
	*RemoteCallbacks
	Bare bool
	IgnoreCertErrors bool
	RemoteName string
	CheckoutBranch string
}

func Clone(url string, path string, options *CloneOptions) (*Repository, error) {
	repo := new(Repository)

      	curl := C.CString(url)
      	defer C.free(unsafe.Pointer(curl))

      	cpath := C.CString(path)
      	defer C.free(unsafe.Pointer(cpath))

	var copts C.git_clone_options
	populateCloneOptions(&copts, options)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ret := C.git_clone(&repo.ptr, curl, cpath, &copts)
	if ret < 0 {
		return nil, LastError()
	}

	runtime.SetFinalizer(repo, (*Repository).Free)
	return repo, nil
}

func populateCloneOptions(ptr *C.git_clone_options, opts *CloneOptions) {
	*ptr = C.git_clone_options_init()
	if opts == nil {
		return
	}
	populateCheckoutOpts(&ptr.checkout_opts, opts.CheckoutOpts)
	populateRemoteCallbacks(&ptr.remote_callbacks, opts.RemoteCallbacks)
	if opts.Bare {
		ptr.bare = 1
	} else {
		ptr.bare = 0
	}
	if opts.IgnoreCertErrors {
		ptr.ignore_cert_errors = 1
	} else {
		ptr.ignore_cert_errors = 0
	}
}

