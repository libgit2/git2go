package git

/*
#include <git2.h>
#include <git2/sys/mempack.h>
#include <git2/pack.h>

extern int git_mempack_new(git_odb_backend **out);
extern int git_mempack_dump(git_buf *pack, git_repository *repo, git_odb_backend *backend);
extern int git_mempack_reset(git_odb_backend *backend);
extern void _go_git_odb_backend_free(git_odb_backend *backend);
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// Mempack is a custom ODB backend that permits packing object in-memory.
type Mempack struct {
	doNotCompare
	ptr *C.git_odb_backend
}

// NewMempack creates a new mempack instance and registers it to the ODB.
func NewMempack(odb *Odb) (mempack *Mempack, err error) {
	mempack = new(Mempack)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_mempack_new(&mempack.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	ret = C.git_odb_add_backend(odb.ptr, mempack.ptr, C.int(999))
	runtime.KeepAlive(odb)
	if ret < 0 {
		// Since git_odb_add_alternate() takes ownership of the ODB backend, the
		// only case in which we free the mempack's memory is if it fails to be
		// added to the ODB.
		C._go_git_odb_backend_free(mempack.ptr)
		return nil, MakeGitError(ret)
	}

	return mempack, nil
}

// Dump dumps all the queued in-memory writes to a packfile.
//
// It is the caller's responsibility to ensure that the generated packfile is
// available to the repository (e.g. by writing it to disk, or doing something
// crazy like distributing it across several copies of the repository over a
// network).
//
// Once the generated packfile is available to the repository, call
// Mempack.Reset to cleanup the memory store.
//
// Calling Mempack.Reset before the packfile has been written to disk will
// result in an inconsistent repository (the objects in the memory store won't
// be accessible).
func (mempack *Mempack) Dump(repository *Repository) ([]byte, error) {
	buf := C.git_buf{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_mempack_dump(&buf, repository.ptr, mempack.ptr)
	runtime.KeepAlive(repository)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	defer C.git_buf_dispose(&buf)

	return C.GoBytes(unsafe.Pointer(buf.ptr), C.int(buf.size)), nil
}

// Reset resets the memory packer by clearing all the queued objects.
//
// This assumes that Mempack.Dump has been called before to store all the
// queued objects into a single packfile.
func (mempack *Mempack) Reset() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_mempack_reset(mempack.ptr)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

// WriteThinPack writes a thin packfile with the objects in the memory store
// into the given packbuilder. A thin packfile does not contain its transitive
// closure of references. This does not reset the in-memory object database.
func (mempack *Mempack) WriteThinPack(pb *Packbuilder) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_mempack_write_thin_pack(mempack.ptr, pb.ptr)
	runtime.KeepAlive(mempack)
	runtime.KeepAlive(pb)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

// ObjectCount returns the total number of objects in the mempack.
func (mempack *Mempack) ObjectCount() (uint, error) {
	var count C.size_t

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_mempack_object_count(&count, mempack.ptr)
	runtime.KeepAlive(mempack)
	if ret < 0 {
		return 0, MakeGitError(ret)
	}
	return uint(count), nil
}
