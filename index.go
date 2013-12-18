package git

/*
#include <git2.h>
#include <git2/errors.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type Index struct {
	ptr *C.git_index
}

func newIndexFromC(ptr *C.git_index) *Index {
	idx := &Index{ptr}
	runtime.SetFinalizer(idx, (*Index).Free)
	return idx
}

func (v *Index) AddByPath(path string) error {
	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_index_add_bypath(v.ptr, cstr)
	if ret < 0 {
		return LastError()
	}

	return nil
}

func (v *Index) WriteTree() (*Oid, error) {
	oid := new(Oid)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_index_write_tree(oid.toC(), v.ptr)
	if ret < 0 {
		return nil, LastError()
	}

	return oid, nil
}

func (v *Index) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_index_free(v.ptr)
}
