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

func (v *Blob) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_object_free(v.ptr)
}

func (v *Blob) Contents() []byte {
	size := C.int(C.git_blob_rawsize(v.ptr))
	buffer := unsafe.Pointer(C.git_blob_rawcontent(v.ptr))
	return C.GoBytes(buffer, size)
}

