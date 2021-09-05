package git

/*
#include <git2.h>
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
	doNotCompare
	ptr *C.git_packbuilder
	r   *Repository
}

func (repo *Repository) NewPackbuilder() (*Packbuilder, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_packbuilder
	ret := C.git_packbuilder_new(&ptr, repo.ptr)
	if ret != 0 {
		return nil, MakeGitError(ret)
	}
	return newPackbuilderFromC(ptr, repo), nil
}

func newPackbuilderFromC(ptr *C.git_packbuilder, r *Repository) *Packbuilder {
	pb := &Packbuilder{ptr: ptr, r: r}
	runtime.SetFinalizer(pb, (*Packbuilder).Free)
	return pb
}

func (pb *Packbuilder) Free() {
	runtime.SetFinalizer(pb, nil)
	C.git_packbuilder_free(pb.ptr)
}

func (pb *Packbuilder) Insert(id *Oid, name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_packbuilder_insert(pb.ptr, id.toC(), cname)
	runtime.KeepAlive(pb)
	runtime.KeepAlive(id)
	if ret != 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (pb *Packbuilder) InsertCommit(id *Oid) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_packbuilder_insert_commit(pb.ptr, id.toC())
	runtime.KeepAlive(pb)
	runtime.KeepAlive(id)
	if ret != 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (pb *Packbuilder) InsertTree(id *Oid) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_packbuilder_insert_tree(pb.ptr, id.toC())
	runtime.KeepAlive(pb)
	runtime.KeepAlive(id)
	if ret != 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (pb *Packbuilder) InsertWalk(walk *RevWalk) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_packbuilder_insert_walk(pb.ptr, walk.ptr)
	runtime.KeepAlive(pb)
	runtime.KeepAlive(walk)
	if ret != 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (pb *Packbuilder) ObjectCount() uint32 {
	ret := uint32(C.git_packbuilder_object_count(pb.ptr))
	runtime.KeepAlive(pb)
	return ret
}

func (pb *Packbuilder) WriteToFile(name string, mode os.FileMode) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_packbuilder_write(pb.ptr, cname, C.uint(mode.Perm()), nil, nil)
	runtime.KeepAlive(pb)
	if ret != 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (pb *Packbuilder) Write(w io.Writer) error {
	return pb.ForEach(func(slice []byte) error {
		_, err := w.Write(slice)
		return err
	})
}

func (pb *Packbuilder) Written() uint32 {
	ret := uint32(C.git_packbuilder_written(pb.ptr))
	runtime.KeepAlive(pb)
	return ret
}

type PackbuilderForeachCallback func([]byte) error
type packbuilderCallbackData struct {
	callback    PackbuilderForeachCallback
	errorTarget *error
}

//export packbuilderForEachCallback
func packbuilderForEachCallback(buf unsafe.Pointer, size C.size_t, handle unsafe.Pointer) C.int {
	payload := pointerHandles.Get(handle)
	data, ok := payload.(*packbuilderCallbackData)
	if !ok {
		panic("could not get packbuilder CB data")
	}

	slice := C.GoBytes(buf, C.int(size))

	err := data.callback(slice)
	if err != nil {
		*data.errorTarget = err
		return C.int(ErrorCodeUser)
	}

	return C.int(ErrorCodeOK)
}

// ForEach repeatedly calls the callback with new packfile data until
// there is no more data or the callback returns an error
func (pb *Packbuilder) ForEach(callback PackbuilderForeachCallback) error {
	var err error
	data := packbuilderCallbackData{
		callback:    callback,
		errorTarget: &err,
	}
	handle := pointerHandles.Track(&data)
	defer pointerHandles.Untrack(handle)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C._go_git_packbuilder_foreach(pb.ptr, handle)
	runtime.KeepAlive(pb)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}
