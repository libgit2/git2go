package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
#include <git2/pack.h>
#include <stdlib.h>

extern int _go_git_packbuilder_foreach(git_packbuilder *pb, void *payload);
*/
import "C"
import (
	"io"
	"runtime"
	"unsafe"
)

type Packbuilder struct {
	ptr *C.git_packbuilder
}

func (repo *Repository) NewPackbuilder() (*Packbuilder, error) {
	builder := &Packbuilder{}
	ret := C.git_packbuilder_new(&builder.ptr, repo.ptr)
	if ret != 0 {
		return nil, LastError()
	}
	runtime.SetFinalizer(builder, (*Packbuilder).Free)
	return builder, nil
}

func (pb *Packbuilder) Free() {
	runtime.SetFinalizer(pb, nil)
	C.git_packbuilder_free(pb.ptr)
}

func (pb *Packbuilder) Insert(id *Oid, name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ret := C.git_packbuilder_insert(pb.ptr, id.toC(), cname)
	if ret != 0 {
		return LastError()
	}
	return nil
}

func (pb *Packbuilder) InsertCommit(id *Oid) error {
	ret := C.git_packbuilder_insert_commit(pb.ptr, id.toC())
	if ret != 0 {
		return LastError()
	}
	return nil
}

func (pb *Packbuilder) InsertTree(id *Oid) error {
	ret := C.git_packbuilder_insert_tree(pb.ptr, id.toC())
	if ret != 0 {
		return LastError()
	}
	return nil
}

func (pb *Packbuilder) ObjectCount() uint32 {
	return uint32(C.git_packbuilder_object_count(pb.ptr))
}

func (pb *Packbuilder) WriteToFile(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ret := C.git_packbuilder_write(pb.ptr, cname, nil, nil)
	if ret != 0 {
		return LastError()
	}
	return nil
}

func (pb *Packbuilder) Write(w io.Writer) error {
	ch := pb.ForEach()
	for slice := range ch {
		_, err := w.Write(slice)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pb *Packbuilder) Written() uint32 {
	return uint32(C.git_packbuilder_written(pb.ptr))
}

//export packbuilderForEachCb
func packbuilderForEachCb(buf unsafe.Pointer, size C.size_t, payload unsafe.Pointer) int {
	ch := *(*chan []byte)(payload)

	slice := C.GoBytes(buf, C.int(size))
	ch <- slice
	return 0
}

func (pb *Packbuilder) forEachWrap(ch chan []byte) {
	C._go_git_packbuilder_foreach(pb.ptr, unsafe.Pointer(&ch))
	close(ch)
}

func (pb *Packbuilder) ForEach() chan []byte {
	ch := make(chan []byte, 0)
	go pb.forEachWrap(ch)
	return ch
}
