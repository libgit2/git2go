package git

/*
#include <git2.h>
#include <git2/errors.h>

static git_remote_callbacks git_remote_callbacks_init() {
	git_remote_callbacks ret = GIT_REMOTE_CALLBACKS_INIT;
	return ret;
}

extern void _setup_callbacks(git_remote_callbacks *callbacks);

*/
import "C"
import (
	"unsafe"
)

type RemoteCompletion uint
const (
	RemoteCompletionDownload RemoteCompletion = C.GIT_REMOTE_COMPLETION_DOWNLOAD
	RemoteCompletionIndexing		  = C.GIT_REMOTE_COMPLETION_INDEXING
	RemoteCompletionError			  = C.GIT_REMOTE_COMPLETION_ERROR
)

type ProgressCallback func(str string) int
type CompletionCallback func(RemoteCompletion) int
type CredentialsCallback func(url string, username_from_url string, allowed_types uint) int // FIXME
type TransferProgressCallback func() int // FIXME
type UpdateTipsCallback func(refname string, a *Oid, b *Oid) int

//export progressCallback
func progressCallback(_str *C.char, _len C.int, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	str := C.GoStringN(_str, _len)	
	return callbacks.ProgressCallback(str)
}

//export completionCallback
func completionCallback(completion_type C.git_remote_completion_type, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	return callbacks.CompletionCallback((RemoteCompletion)(completion_type))
}

//export credentialsCallback
func credentialsCallback(_cred **C.git_cred, _url *C.char, _username_from_url *C.char, allowed_types uint, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	//cred := C.GoString(_cred)
	url := C.GoString(_url)
	username_from_url := C.GoString(_username_from_url)
	return callbacks.CredentialsCallback(url, username_from_url, allowed_types)
}

//export transferProgressCallback
func transferProgressCallback(stats C.git_transfer_progress, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	return callbacks.TransferProgressCallback()
}

//export updateTipsCallback
func updateTipsCallback(_refname *C.char, _a *C.git_oid, _b *C.git_oid, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	refname := C.GoString(_refname)
	a := newOidFromC(_a)
	b := newOidFromC(_b)
	return callbacks.UpdateTipsCallback(refname, a, b)
}

type RemoteCallbacks struct {
	ProgressCallback
	CompletionCallback
	CredentialsCallback
	TransferProgressCallback
	UpdateTipsCallback
}

func populateRemoteCallbacks(ptr *C.git_remote_callbacks, callbacks *RemoteCallbacks) {
	*ptr = C.git_remote_callbacks_init()
	if callbacks == nil {
		return
	}
	C._setup_callbacks(ptr)
	ptr.payload = unsafe.Pointer(callbacks)
}
