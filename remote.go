package git

/*
#include <git2.h>
#include <git2/errors.h>

extern void _go_git_setup_callbacks(git_remote_callbacks *callbacks);
extern git_remote_callbacks _go_git_remote_callbacks_init();

*/
import "C"
import "unsafe"

type TransferProgress struct {
	ptr *C.git_transfer_progress
}

type RemoteCompletion uint
const (
	RemoteCompletionDownload RemoteCompletion = C.GIT_REMOTE_COMPLETION_DOWNLOAD
	RemoteCompletionIndexing		  = C.GIT_REMOTE_COMPLETION_INDEXING
	RemoteCompletionError			  = C.GIT_REMOTE_COMPLETION_ERROR
)

type ProgressCallback func(str string) int
type CompletionCallback func(RemoteCompletion) int
type CredentialsCallback func(url string, username_from_url string, allowed_types CredType) (int, Cred)
type TransferProgressCallback func(stats TransferProgress) int
type UpdateTipsCallback func(refname string, a *Oid, b *Oid) int

type RemoteCallbacks struct {
	ProgressCallback
	CompletionCallback
	CredentialsCallback
	TransferProgressCallback
	UpdateTipsCallback
}

func populateRemoteCallbacks(ptr *C.git_remote_callbacks, callbacks *RemoteCallbacks) {
	*ptr = C._go_git_remote_callbacks_init()
	if callbacks == nil {
		return
	}
	C._go_git_setup_callbacks(ptr)
	ptr.payload = unsafe.Pointer(callbacks)
}

//export progressCallback
func progressCallback(_str *C.char, _len C.int, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	if callbacks.ProgressCallback == nil {
		return 0
	}
	str := C.GoStringN(_str, _len)
	return callbacks.ProgressCallback(str)
}

//export completionCallback
func completionCallback(completion_type C.git_remote_completion_type, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	if callbacks.CompletionCallback == nil {
		return 0
	}
	return callbacks.CompletionCallback((RemoteCompletion)(completion_type))
}

//export credentialsCallback
func credentialsCallback(_cred **C.git_cred, _url *C.char, _username_from_url *C.char, allowed_types uint, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	if callbacks.CredentialsCallback == nil {
		return 0
	}
	url := C.GoString(_url)
	username_from_url := C.GoString(_username_from_url)
	ret, cred := callbacks.CredentialsCallback(url, username_from_url, (CredType)(allowed_types))
	if gcred, ok := cred.(gitCred); ok {
		*_cred = gcred.ptr
	}
	return ret
}

//export transferProgressCallback
func transferProgressCallback(stats *C.git_transfer_progress, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	if callbacks.TransferProgressCallback == nil {
		return 0
	}
	return callbacks.TransferProgressCallback(TransferProgress{stats})
}

//export updateTipsCallback
func updateTipsCallback(_refname *C.char, _a *C.git_oid, _b *C.git_oid, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	if callbacks.UpdateTipsCallback == nil {
		return 0
	}
	refname := C.GoString(_refname)
	a := newOidFromC(_a)
	b := newOidFromC(_b)
	return callbacks.UpdateTipsCallback(refname, a, b)
}

func (o TransferProgress) TotalObjects() uint {
	return uint(o.ptr.total_objects)
}

func (o TransferProgress) IndexedObjects() uint {
	return uint(o.ptr.indexed_objects)
}

func (o TransferProgress) ReceivedObjects() uint {
	return uint(o.ptr.received_objects)
}

func (o TransferProgress) LocalObjects() uint {
	return uint(o.ptr.local_objects)
}

func (o TransferProgress) TotalDeltas() uint {
	return uint(o.ptr.total_deltas)
}

func (o TransferProgress) ReceivedBytes() uint {
	return uint(o.ptr.received_bytes)
}
