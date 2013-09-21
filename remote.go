package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>

extern int _go_git_remote_ls(git_remote *remote, void *payload);
extern int _go_git_remote_set_callbacks(git_remote *remote, void *payload);
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type RemoteDirection int

const (
	RemoteDirectionFetch RemoteDirection = C.GIT_DIRECTION_FETCH
	RemoteDirectionPush                  = C.GIT_DIRECTION_PUSH
)

type AutotagOption int

const (
	AutotagAuto AutotagOption = C.GIT_REMOTE_DOWNLOAD_TAGS_AUTO
	AutotagNone               = C.GIT_REMOTE_DOWNLOAD_TAGS_NONE
	AutotagAll                = C.GIT_REMOTE_DOWNLOAD_TAGS_ALL
)

type ProgressCb func([]byte) int
type TransferProgressCb func(*TransferProgress) int
type UpdateTipsCb func(string, *Oid, *Oid) int

type Remote struct {
	Name string
	Url  string

	// callbacks
	Progress   ProgressCb
	TransferProgress TransferProgressCb
	UpdateTips UpdateTipsCb

	ptr  *C.git_remote
}

func (r *Remote) Connect(direction RemoteDirection) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_connect(r.ptr, C.git_direction(direction))
	if ret < 0 {
		return LastError()
	}

	return nil
}

func (r *Remote) IsConnected() bool {
	return C.git_remote_connected(r.ptr) != 0
}

func (r *Remote) Disconnect() {
	C.git_remote_disconnect(r.ptr)
}

func (r *Remote) Autotag() AutotagOption {
	return AutotagOption(C.git_remote_autotag(r.ptr))
}

func (r *Remote) SetAutotag(opt AutotagOption) {
	C.git_remote_set_autotag(r.ptr, C.git_remote_autotag_option_t(opt))
}

func (r *Remote) Stop() {
	C.git_remote_stop(r.ptr)
}

func (r *Remote) Save() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_save(r.ptr)
	if ret < 0 {
		return LastError()
	}

	return nil
}

func (r *Remote) Ls() ([]*RemoteHead, error) {
	var data headlistData

	ret := C._go_git_remote_ls(r.ptr, unsafe.Pointer(&data))
	if ret < 0 {
		return nil, LastError()
	}

	return data.slice, nil
}

func (r *Remote) Download() error {
	ret := C.git_remote_download(r.ptr)
	if ret < 0 {
		LastError()
	}

	return nil
}

func (r *Remote) Fetch() error {
	ret := C.git_remote_fetch(r.ptr)
	if ret < 0 {
		LastError()
	}

	return nil
}

//export remoteProgress
func remoteProgress(str *C.char, length C.int, data unsafe.Pointer) int {
	remote := (*Remote)(data)
	if remote.Progress != nil {
		return remote.Progress(C.GoBytes(unsafe.Pointer(str), length))
	}

	return 0
}

//export remoteTransferProgress
func remoteTransferProgress(ptr *C.git_transfer_progress, data unsafe.Pointer) int {
	remote := (*Remote)(data)
	if remote.TransferProgress != nil {
		return remote.TransferProgress(newTransferProgressFromC(ptr))
	}

	return 0
}

//export remoteUpdateTips
func remoteUpdateTips(str *C.char, a, b *C.git_oid, data unsafe.Pointer) int {
	remote := (*Remote)(data)
	if remote.UpdateTips != nil {
		var goa, gob Oid
		CopyOid(&goa, a)
		CopyOid(&gob, b)
		return remote.UpdateTips(C.GoString(str), &goa, &gob)
	}

	return 0
}

type headlistData struct {
	slice []*RemoteHead
}

//export remoteHeadlistCb
func remoteHeadlistCb(rhead *C.git_remote_head, dataptr unsafe.Pointer) int {
	data := (*headlistData)(dataptr)

	head := newRemoteHeadFromC(rhead)
	data.slice = append(data.slice, head)

	return 0
}

func (r *Remote) Free() {
	runtime.SetFinalizer(r, nil)
	C.git_remote_free(r.ptr)
}

func newRemoteFromC(ptr *C.git_remote) *Remote {
	remote := &Remote{
		ptr:  ptr,
		Name: C.GoString(C.git_remote_name(ptr)),
		Url:  C.GoString(C.git_remote_url(ptr)),
	}

	// allways set the callbacks, we'll decide whether to call
	// them once we're back in go-land
	C._go_git_remote_set_callbacks(remote.ptr, unsafe.Pointer(remote))
	runtime.SetFinalizer(remote, (*Remote).Free)

	return remote
}

// transfer progress

type TransferProgress struct {
	TotalObjects    uint
	IndexedObjects  uint
	ReceivedObjects uint
	ReceivedBytes   uint64
}

func newTransferProgressFromC(ptr *C.git_transfer_progress) *TransferProgress {
	return &TransferProgress{
		TotalObjects:    uint(ptr.total_objects),
		IndexedObjects:  uint(ptr.indexed_objects),
		ReceivedObjects: uint(ptr.received_objects),
		ReceivedBytes:   uint64(ptr.received_bytes),
	}
}

// remote heads

// RemoteHead represents a reference available in the remote repository.
type RemoteHead struct {
	Local bool
	Oid   Oid
	Loid  Oid
	Name  string
}

func newRemoteHeadFromC(ptr *C.git_remote_head) *RemoteHead {
	head := &RemoteHead {
		Local: ptr.local != 0,
		Name: C.GoString(ptr.name),
	}

	CopyOid(&head.Oid, &ptr.oid)
	CopyOid(&head.Loid, &ptr.loid)

	return head
}

// These belong to the git_remote namespace but don't require any remote

func UrlIsValid(url string) bool {
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))

	return C.git_remote_valid_url(curl) != 0
}


func UrlIsSupported(url string) bool {
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))

	return C.git_remote_supported_url(curl) != 0
}
