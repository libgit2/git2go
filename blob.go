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

type Blob struct {
	ptr *C.git_object
}

// Id() and Type() satisfy Object
func (v *Blob) Id() *Oid {
	return newOidFromC(C.git_blob_id(v.ptr))
}

func (v *Blob) Type() int {
	return OBJ_BLOB
}

func (v *Blob) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_object_free(v.ptr)
}

func newBlobFromC(ptr *C.git_blob) *Blob {
	blob := &Blob{ptr}
	runtime.SetFinalizer(blob, (*Blob).Free)

	return blob
}

func (v *Blob) Size() int64 {
	return int64(C.git_blob_rawsize(v.ptr))
}

func (v *Blob) Contents() []byte {
	size := C.int(C.git_blob_rawsize(v.ptr))
	buffer := unsafe.Pointer(C.git_blob_rawcontent(v.ptr))
	return C.GoBytes(buffer, size)
}

