package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
*/
import "C"
import (
	"fmt"
	"runtime"
	"time"
	"unsafe"
)

type Index struct {
	ptr *C.git_index
}

type IndexEntry struct {
	Ctime time.Time
	Mtime time.Time
	Mode  uint
	Uid   uint
	Gid   uint
	Size  uint
	Oid   *Oid
	Path  string
}

func newIndexFromC(ptr *C.git_index) *Index {
	idx := &Index{ptr}
	runtime.SetFinalizer(idx, (*Index).Free)
	return idx
}

func (v *Index) AddByPath(path string) error {
	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))

	ret := C.git_index_add_bypath(v.ptr, cstr)
	if ret < 0 {
		return LastError()
	}

	return nil
}

func (v *Index) WriteTree() (*Oid, error) {
	oid := new(Oid)
	ret := C.git_index_write_tree(oid.toC(), v.ptr)
	if ret < 0 {
		return nil, LastError()
	}

	return oid, nil
}

func (v *Index) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_index_free(v.ptr)
}

func (v *Index) EntryCount() uint {
	return uint(C.git_index_entrycount(v.ptr))
}

func newIndexEntryFromC(entry *C.git_index_entry) *IndexEntry {
	return &IndexEntry{
		time.Unix(int64(entry.ctime.seconds), int64(entry.ctime.nanoseconds)),
		time.Unix(int64(entry.mtime.seconds), int64(entry.mtime.nanoseconds)),
		uint(entry.mode),
		uint(entry.uid),
		uint(entry.gid),
		uint(entry.file_size),
		newOidFromC(&entry.oid),
		C.GoString(entry.path),
	}
}

func (v *Index) EntryByIndex(index uint) (*IndexEntry, error) {
	centry := C.git_index_get_byindex(v.ptr, C.size_t(index))
	if centry == nil {
		return nil, fmt.Errorf("Index out of Bounds")
	}
	return newIndexEntryFromC(centry), nil
}
