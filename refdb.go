package git

/*
#include <git2.h>
#include <git2/errors.h>
#include <git2/sys/refdb_backend.h>

extern void _go_git_refdb_backend_free(git_refdb_backend *backend);
*/
import "C"
import (
	"runtime"
)

type Refdb struct {
	ptr *C.git_refdb
}

type RefdbBackend struct {
	ptr *C.git_refdb_backend
}

func (v *Repository) NewRefdb() (refdb *Refdb, err error) {
	refdb = new(Refdb)

	ret := C.git_refdb_new(&refdb.ptr, v.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(refdb, (*Refdb).Free)
	return refdb, nil
}

func NewRefdbBackendFromC(ptr *C.git_refdb_backend) (backend *RefdbBackend) {
	backend = &RefdbBackend{ptr}
	return backend
}

func (v *Refdb) SetBackend(backend *RefdbBackend) (err error) {
	ret := C.git_refdb_set_backend(v.ptr, backend.ptr)
	if ret < 0 {
		backend.Free()
		return MakeGitError(ret)
	}
	return nil
}

func (v *RefdbBackend) Free() {
	C._go_git_refdb_backend_free(v.ptr)
}
