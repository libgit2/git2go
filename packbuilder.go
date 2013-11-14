package git

/*
#include <git2.h>
#include <git2/errors.h>
#include <git2/pack.h>
#include <stdlib.h>

extern int _go_git_packbuilder_foreach(git_packbuilder *pb, void *payload);
*/
import "C"
import (
	"io"
	"os"
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
		return nil, makeError(ret)
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
	return makeError(C.git_packbuilder_insert(pb.ptr, id.toC(), cname))
}

func (pb *Packbuilder) InsertCommit(id *Oid) error {
	return makeError(C.git_packbuilder_insert_commit(pb.ptr, id.toC()))
}

func (pb *Packbuilder) InsertTree(id *Oid) error {
	return makeError(C.git_packbuilder_insert_tree(pb.ptr, id.toC()))
}

func (pb *Packbuilder) ObjectCount() uint32 {
	return uint32(C.git_packbuilder_object_count(pb.ptr))
}

func (pb *Packbuilder) WriteToFile(name string, mode os.FileMode) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ret := C.git_packbuilder_write(pb.ptr, cname, C.uint(mode.Perm()), nil, nil)
	if ret != 0 {
		return makeError(ret)
	}
	return nil
}

func (pb *Packbuilder) Write(w io.Writer) error {
	ch, stop := pb.ForEach()
	for slice := range ch {
		_, err := w.Write(slice)
		if err != nil {
			close(stop)
			return err
		}
	}
	return nil
}

func (pb *Packbuilder) Written() uint32 {
	return uint32(C.git_packbuilder_written(pb.ptr))
}

type packbuilderCbData struct {
	ch chan<- []byte
	stop <-chan bool
}

//export packbuilderForEachCb
func packbuilderForEachCb(buf unsafe.Pointer, size C.size_t, payload unsafe.Pointer) int {
	data := (*packbuilderCbData)(payload)
	ch := data.ch
	stop := data.stop

	slice := C.GoBytes(buf, C.int(size))
	select {
	case <- stop:
		return -1
	case ch <- slice:
	}

	return 0
}

func (pb *Packbuilder) forEachWrap(data *packbuilderCbData) {
	C._go_git_packbuilder_foreach(pb.ptr, unsafe.Pointer(data))
	close(data.ch)
}

// Foreach sends the packfile as slices through the "data" channel. If
// you want to stop the pack-building process (e.g. there's an error
// writing to the output), close or write a value into the "stop"
// channel.
func (pb *Packbuilder) ForEach() (<-chan []byte, chan<- bool) {
	ch := make(chan []byte)
	stop := make(chan bool)
	data := packbuilderCbData{ch, stop}
	go pb.forEachWrap(&data)
	return ch, stop
}
