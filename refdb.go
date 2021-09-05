package git

/*
#include <git2.h>
#include <git2/sys/refdb_backend.h>

extern void _go_git_refdb_backend_free(git_refdb_backend *backend);
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type Refdb struct {
	doNotCompare
	ptr *C.git_refdb
	r   *Repository
}

type RefdbBackend struct {
	doNotCompare
	ptr *C.git_refdb_backend
}

func (v *Repository) NewRefdb() (refdb *Refdb, err error) {
	var ptr *C.git_refdb

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_refdb_new(&ptr, v.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	refdb = &Refdb{ptr: ptr, r: v}
	runtime.SetFinalizer(refdb, (*Refdb).Free)
	return refdb, nil
}

func NewRefdbBackendFromC(ptr unsafe.Pointer) (backend *RefdbBackend) {
	backend = &RefdbBackend{ptr: (*C.git_refdb_backend)(ptr)}
	return backend
}

func (v *Refdb) SetBackend(backend *RefdbBackend) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_refdb_set_backend(v.ptr, backend.ptr)
	runtime.KeepAlive(v)
	runtime.KeepAlive(backend)
	if ret < 0 {
		backend.Free()
		return MakeGitError(ret)
	}
	return nil
}

func (v *RefdbBackend) Free() {
	runtime.SetFinalizer(v, nil)
	C._go_git_refdb_backend_free(v.ptr)
}
