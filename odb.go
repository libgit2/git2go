package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>

extern int _go_git_odb_foreach(git_odb *db, void *payload);
*/
import "C"
import (
	"unsafe"
	"reflect"
	"runtime"
)

type Odb struct {
	ptr *C.git_odb
}

func (v *Odb) Exists(oid *Oid) bool {
	ret := C.git_odb_exists(v.ptr, oid.toC())
	return ret != 0
}

func (v *Odb) Write(data []byte, otype ObjectType) (oid *Oid, err error) {
	oid = new(Oid)
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	ret := C.git_odb_write(oid.toC(), v.ptr, unsafe.Pointer(hdr.Data), C.size_t(hdr.Len), C.git_otype(otype))

	if ret < 0 {
		err = MakeGitError(ret)
	}

	return
}

func (v *Odb) Read(oid *Oid) (obj *OdbObject, err error) {
	obj = new(OdbObject)
	ret := C.git_odb_read(&obj.ptr, v.ptr, oid.toC())
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(obj, (*OdbObject).Free)
	return
}

//export odbForEachCb
func odbForEachCb(id *C.git_oid, payload unsafe.Pointer) int {
	ch := *(*chan *Oid)(payload)
	oid := newOidFromC(id)
	// Because the channel is unbuffered, we never read our own data. If ch is
	// readable, the user has sent something on it, which means we should
	// abort.
	select {
	case ch <- oid:
	case <-ch:
			return -1
	}
	return 0;
}

func (v *Odb) forEachWrap(ch chan *Oid) {
	C._go_git_odb_foreach(v.ptr, unsafe.Pointer(&ch))
	close(ch)
}

func (v *Odb) ForEach() chan *Oid {
	ch := make(chan *Oid, 0)
	go v.forEachWrap(ch)
	return ch
}

type OdbObject struct {
	ptr *C.git_odb_object
}

func (v *OdbObject) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_odb_object_free(v.ptr)
}

func (object *OdbObject) Id() (oid *Oid) {
	return newOidFromC(C.git_odb_object_id(object.ptr))
}

func (object *OdbObject) Len() (len uint64) {
	return uint64(C.git_odb_object_size(object.ptr))
}

func (object *OdbObject) Data() (data []byte) {
	var c_blob unsafe.Pointer = C.git_odb_object_data(object.ptr)
	var blob []byte

	len := int(C.git_odb_object_size(object.ptr))

	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&blob)))
	sliceHeader.Cap = len
	sliceHeader.Len = len
	sliceHeader.Data = uintptr(c_blob)

	return blob
}
