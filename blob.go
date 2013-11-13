package git

/*
#include <git2.h>
#include <git2/errors.h>
*/
import "C"
import (
	"unsafe"
)

type Blob struct {
	gitObject
}

func (v Blob) Size() int64 {
	return int64(C.git_blob_rawsize(v.ptr))
}

func (v Blob) Contents() []byte {
	size := C.int(C.git_blob_rawsize(v.ptr))
	buffer := unsafe.Pointer(C.git_blob_rawcontent(v.ptr))
	return C.GoBytes(buffer, size)
}

