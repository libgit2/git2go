package git

/*
#include <git2.h>

extern void _go_git_populate_clone_callbacks(git_clone_options *opts);
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type RemoteCreateCallback func(repo *Repository, name, url string) (*Remote, error)

type CloneOptions struct {
	CheckoutOptions      CheckoutOptions
	FetchOptions         FetchOptions
	Bare                 bool
	CheckoutBranch       string
	RemoteCreateCallback RemoteCreateCallback
}

func Clone(url string, path string, options *CloneOptions) (*Repository, error) {
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	var err error
	cOptions := populateCloneOptions(&C.git_clone_options{}, options, &err)
	defer freeCloneOptions(cOptions)

	if len(options.CheckoutBranch) != 0 {
		cOptions.checkout_branch = C.CString(options.CheckoutBranch)
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_repository
	ret := C.git_clone(&ptr, curl, cpath, cOptions)

	if ret == C.int(ErrorCodeUser) && err != nil {
		return nil, err
	}
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newRepositoryFromC(ptr), nil
}

//export remoteCreateCallback
func remoteCreateCallback(
	out **C.git_remote,
	crepo *C.git_repository,
	cname, curl *C.char,
	handle unsafe.Pointer,
) C.int {
	name := C.GoString(cname)
	url := C.GoString(curl)
	repo := newRepositoryFromC(crepo)
	repo.weak = true
	defer repo.Free()

	data, ok := pointerHandles.Get(handle).(*cloneCallbackData)
	if !ok {
		panic("invalid remote create callback")
	}

	remote, err := data.options.RemoteCreateCallback(repo, name, url)

	if err != nil {
		*data.errorTarget = err
		return C.int(ErrorCodeUser)
	}
	if remote == nil {
		panic("no remote created by callback")
	}

	*out = remote.ptr

	// clear finalizer as the calling C function will
	// free the remote itself
	runtime.SetFinalizer(remote, nil)
	remote.repo.Remotes.untrackRemote(remote)

	return C.int(ErrorCodeOK)
}

type cloneCallbackData struct {
	options     *CloneOptions
	errorTarget *error
}

func populateCloneOptions(copts *C.git_clone_options, opts *CloneOptions, errorTarget *error) *C.git_clone_options {
	C.git_clone_options_init(copts, C.GIT_CLONE_OPTIONS_VERSION)
	if opts == nil {
		return nil
	}
	populateCheckoutOptions(&copts.checkout_opts, &opts.CheckoutOptions, errorTarget)
	populateFetchOptions(&copts.fetch_opts, &opts.FetchOptions, errorTarget)
	copts.bare = cbool(opts.Bare)

	if opts.RemoteCreateCallback != nil {
		data := &cloneCallbackData{
			options:     opts,
			errorTarget: errorTarget,
		}
		// Go v1.1 does not allow to assign a C function pointer
		C._go_git_populate_clone_callbacks(copts)
		copts.remote_cb_payload = pointerHandles.Track(data)
	}

	return copts
}

func freeCloneOptions(copts *C.git_clone_options) {
	if copts == nil {
		return
	}

	freeCheckoutOptions(&copts.checkout_opts)
	freeFetchOptions(&copts.fetch_opts)

	if copts.remote_cb_payload != nil {
		pointerHandles.Untrack(copts.remote_cb_payload)
	}

	C.free(unsafe.Pointer(copts.checkout_branch))
}
