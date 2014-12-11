package git

/*
#include <git2.h>

extern int _go_git_index_add_all(git_index*, const git_strarray*, unsigned int, void*);
extern int _go_git_index_update_all(git_index*, const git_strarray*, void*);
extern int _go_git_index_remove_all(git_index*, const git_strarray*, void*);

*/
import "C"
import (
	"fmt"
	"runtime"
	"time"
	"unsafe"
)

type IndexMatchedPathCallback func(string, string) int

type IndexAddOpts uint

const (
	IndexAddDefault              IndexAddOpts = C.GIT_INDEX_ADD_DEFAULT
	IndexAddForce                IndexAddOpts = C.GIT_INDEX_ADD_FORCE
	IndexAddDisablePathspecMatch IndexAddOpts = C.GIT_INDEX_ADD_DISABLE_PATHSPEC_MATCH
	IndexAddCheckPathspec        IndexAddOpts = C.GIT_INDEX_ADD_CHECK_PATHSPEC
)

type Index struct {
	ptr *C.git_index
}

type IndexEntry struct {
	Ctime time.Time
	Mtime time.Time
	Mode  Filemode
	Uid   uint
	Gid   uint
	Size  uint
	Id    *Oid
	Path  string
}

func newIndexEntryFromC(entry *C.git_index_entry) *IndexEntry {
	if entry == nil {
		return nil
	}
	return &IndexEntry{
		time.Unix(int64(entry.ctime.seconds), int64(entry.ctime.nanoseconds)),
		time.Unix(int64(entry.mtime.seconds), int64(entry.mtime.nanoseconds)),
		Filemode(entry.mode),
		uint(entry.uid),
		uint(entry.gid),
		uint(entry.file_size),
		newOidFromC(&entry.id),
		C.GoString(entry.path),
	}
}

func populateCIndexEntry(source *IndexEntry, dest *C.git_index_entry) {
	dest.ctime.seconds = C.git_time_t(source.Ctime.Unix())
	dest.ctime.nanoseconds = C.uint(source.Ctime.UnixNano())
	dest.mtime.seconds = C.git_time_t(source.Mtime.Unix())
	dest.mtime.nanoseconds = C.uint(source.Mtime.UnixNano())
	dest.mode = C.uint(source.Mode)
	dest.uid = C.uint(source.Uid)
	dest.gid = C.uint(source.Gid)
	dest.file_size = C.git_off_t(source.Size)
	dest.id = *source.Id.toC()
	dest.path = C.CString(source.Path)
}

func freeCIndexEntry(entry *C.git_index_entry) {
	C.free(unsafe.Pointer(entry.path))
}

func newIndexFromC(ptr *C.git_index) *Index {
	idx := &Index{ptr}
	runtime.SetFinalizer(idx, (*Index).Free)
	return idx
}

// NewIndex allocates a new index. It won't be associated with any
// file on the filesystem or repository
func NewIndex() (*Index, error) {
	var ptr *C.git_index

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.git_index_new(&ptr); err < 0 {
		return nil, MakeGitError(err)
	}

	return &Index{ptr: ptr}, nil
}

// Add adds or replaces the given entry to the index, making a copy of
// the data
func (v *Index) Add(entry *IndexEntry) error {
	var centry C.git_index_entry

	populateCIndexEntry(entry, &centry)
	defer freeCIndexEntry(&centry)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.git_index_add(v.ptr, &centry); err < 0 {
		return MakeGitError(err)
	}

	return nil
}

func (v *Index) AddByPath(path string) error {
	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_index_add_bypath(v.ptr, cstr)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

func (v *Index) AddAll(pathspecs []string, flags IndexAddOpts, callback IndexMatchedPathCallback) error {
	cpathspecs := C.git_strarray{}
	cpathspecs.count = C.size_t(len(pathspecs))
	cpathspecs.strings = makeCStringsFromStrings(pathspecs)
	defer freeStrarray(&cpathspecs)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var cb *IndexMatchedPathCallback
	if callback != nil {
		cb = &callback
	}

	ret := C._go_git_index_add_all(
		v.ptr,
		&cpathspecs,
		C.uint(flags),
		unsafe.Pointer(cb),
	)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (v *Index) UpdateAll(pathspecs []string, callback IndexMatchedPathCallback) error {
	cpathspecs := C.git_strarray{}
	cpathspecs.count = C.size_t(len(pathspecs))
	cpathspecs.strings = makeCStringsFromStrings(pathspecs)
	defer freeStrarray(&cpathspecs)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var cb *IndexMatchedPathCallback
	if callback != nil {
		cb = &callback
	}

	ret := C._go_git_index_update_all(
		v.ptr,
		&cpathspecs,
		unsafe.Pointer(cb),
	)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (v *Index) RemoveAll(pathspecs []string, callback IndexMatchedPathCallback) error {
	cpathspecs := C.git_strarray{}
	cpathspecs.count = C.size_t(len(pathspecs))
	cpathspecs.strings = makeCStringsFromStrings(pathspecs)
	defer freeStrarray(&cpathspecs)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var cb *IndexMatchedPathCallback
	if callback != nil {
		cb = &callback
	}

	ret := C._go_git_index_remove_all(
		v.ptr,
		&cpathspecs,
		unsafe.Pointer(cb),
	)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

//export indexMatchedPathCallback
func indexMatchedPathCallback(cPath, cMatchedPathspec *C.char, payload unsafe.Pointer) int {
	callback := (*IndexMatchedPathCallback)(payload)
	return (*callback)(C.GoString(cPath), C.GoString(cMatchedPathspec))
}

func (v *Index) RemoveByPath(path string) error {
	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_index_remove_bypath(v.ptr, cstr)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

func (v *Index) WriteTreeTo(repo *Repository) (*Oid, error) {
	oid := new(Oid)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_index_write_tree_to(oid.toC(), v.ptr, repo.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return oid, nil
}

func (v *Index) WriteTree() (*Oid, error) {
	oid := new(Oid)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_index_write_tree(oid.toC(), v.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return oid, nil
}

func (v *Index) Write() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_index_write(v.ptr)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

func (v *Index) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_index_free(v.ptr)
}

func (v *Index) EntryCount() uint {
	return uint(C.git_index_entrycount(v.ptr))
}

func (v *Index) EntryByIndex(index uint) (*IndexEntry, error) {
	centry := C.git_index_get_byindex(v.ptr, C.size_t(index))
	if centry == nil {
		return nil, fmt.Errorf("Index out of Bounds")
	}
	return newIndexEntryFromC(centry), nil
}

func (v *Index) HasConflicts() bool {
	return C.git_index_has_conflicts(v.ptr) != 0
}

func (v *Index) CleanupConflicts() {
	C.git_index_conflict_cleanup(v.ptr)
}

func (v *Index) AddConflict(ancestor *IndexEntry, our *IndexEntry, their *IndexEntry) error {

	var cancestor *C.git_index_entry
	var cour *C.git_index_entry
	var ctheir *C.git_index_entry

	if ancestor != nil {
		cancestor = &C.git_index_entry{}
		populateCIndexEntry(ancestor, cancestor)
		defer freeCIndexEntry(cancestor)
	}

	if our != nil {
		cour = &C.git_index_entry{}
		populateCIndexEntry(our, cour)
		defer freeCIndexEntry(cour)
	}

	if their != nil {
		ctheir = &C.git_index_entry{}
		populateCIndexEntry(their, ctheir)
		defer freeCIndexEntry(ctheir)
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_index_conflict_add(v.ptr, cancestor, cour, ctheir)
	if ecode < 0 {
		return MakeGitError(ecode)
	}
	return nil
}

type IndexConflict struct {
	Ancestor *IndexEntry
	Our      *IndexEntry
	Their    *IndexEntry
}

func (v *Index) GetConflict(path string) (IndexConflict, error) {

	var cancestor *C.git_index_entry
	var cour *C.git_index_entry
	var ctheir *C.git_index_entry

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_index_conflict_get(&cancestor, &cour, &ctheir, v.ptr, cpath)
	if ecode < 0 {
		return IndexConflict{}, MakeGitError(ecode)
	}
	return IndexConflict{
		Ancestor: newIndexEntryFromC(cancestor),
		Our:      newIndexEntryFromC(cour),
		Their:    newIndexEntryFromC(ctheir),
	}, nil
}

func (v *Index) RemoveConflict(path string) error {

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_index_conflict_remove(v.ptr, cpath)
	if ecode < 0 {
		return MakeGitError(ecode)
	}
	return nil
}

type IndexConflictIterator struct {
	ptr   *C.git_index_conflict_iterator
	index *Index
}

func newIndexConflictIteratorFromC(index *Index, ptr *C.git_index_conflict_iterator) *IndexConflictIterator {
	i := &IndexConflictIterator{ptr: ptr, index: index}
	runtime.SetFinalizer(i, (*IndexConflictIterator).Free)
	return i
}

func (v *IndexConflictIterator) Index() *Index {
	return v.index
}

func (v *IndexConflictIterator) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_index_conflict_iterator_free(v.ptr)
}

func (v *Index) ConflictIterator() (*IndexConflictIterator, error) {
	var i *C.git_index_conflict_iterator

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_index_conflict_iterator_new(&i, v.ptr)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}
	return newIndexConflictIteratorFromC(v, i), nil
}

func (v *IndexConflictIterator) Next() (IndexConflict, error) {
	var cancestor *C.git_index_entry
	var cour *C.git_index_entry
	var ctheir *C.git_index_entry

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_index_conflict_next(&cancestor, &cour, &ctheir, v.ptr)
	if ecode < 0 {
		return IndexConflict{}, MakeGitError(ecode)
	}
	return IndexConflict{
		Ancestor: newIndexEntryFromC(cancestor),
		Our:      newIndexEntryFromC(cour),
		Their:    newIndexEntryFromC(ctheir),
	}, nil
}
