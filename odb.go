package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
*/
import "C"
import (
	"unsafe"
	"reflect"
	"runtime"
)

var (
	OBJ_ANY = C.GIT_OBJ_ANY
	OBJ_BAD = C.GIT_OBJ_BAD
	OBJ_COMMIT = C.GIT_OBJ_COMMIT
	OBJ_TREE = C.GIT_OBJ_TREE
	OBJ_BLOB = C.GIT_OBJ_BLOB
	OBJ_TAG = C.GIT_OBJ_TAG
)

type Odb struct {
	ptr *C.git_odb
}

func (v *Odb) Exists(oid *Oid) bool {
	ret := C.git_odb_exists(v.ptr, oid.toC())
	return ret != 0
}

func (v *Odb) Write(data []byte, otype int) (oid *Oid, err error) {
	oid = new(Oid)
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	ret := C.git_odb_write(oid.toC(), v.ptr, unsafe.Pointer(hdr.Data), C.size_t(hdr.Len), C.git_otype(otype))

	if ret < 0 {
		err = LastError()
	}

	return
}

func (v *Odb) Read(oid *Oid) (obj *OdbObject, err error) {
	obj = new(OdbObject)
	ret := C.git_odb_read(&obj.ptr, v.ptr, oid.toC())
	if ret < 0 {
		return nil, LastError()
	}

	runtime.SetFinalizer(obj, freeOdbObject)
	return
}

type OdbObject struct {
	ptr *C.git_odb_object
}

func freeOdbObject(obj *OdbObject) {
	C.git_odb_object_free(obj.ptr)
}

func (v *OdbObject) Type() int {
	return int(C.git_odb_object_type(v.ptr))
}

func (v *OdbObject) Size() int64 {
	return int64(C.git_odb_object_size(v.ptr))
}

