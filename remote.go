package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
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

type Remote struct {
	Name string
	Url  string
	ptr  *C.git_remote
}

func (r *Remote) Connect(direction RemoteDirection) error {
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
	ret := C.git_remote_save(r.ptr)
	if ret < 0 {
		return LastError()
	}

	return nil
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

	runtime.SetFinalizer(remote, (*Remote).Free)

	return remote
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
