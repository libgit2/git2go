package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
*/
import "C"
import (
	"unsafe"
	"runtime"
)

type Blob struct {
	ptr *C.git_blob
}

func (o *Blob) Id() *Oid {
	return newOidFromC(C.git_blob_id(o.ptr))
}

func (o *Blob) Type() ObjectType {
	return OBJ_BLOB
}

func (o *Blob) Free() {
	runtime.SetFinalizer(o, nil)
	C.git_blob_free(o.ptr)
}

func (v *Blob) Size() int64 {
	return int64(C.git_blob_rawsize(v.ptr))
}

func (v *Blob) Contents() []byte {
	size := C.int(C.git_blob_rawsize(v.ptr))
	buffer := unsafe.Pointer(C.git_blob_rawcontent(v.ptr))
	return C.GoBytes(buffer, size)
}

