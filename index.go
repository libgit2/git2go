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
	"unsafe"
)

type IndexMatchedPathCallback func(string, string) error
type indexMatchedPathCallbackData struct {
	callback    IndexMatchedPathCallback
	errorTarget *error
}

// IndexAddOption is a set of flags for APIs that add files matching pathspec.
type IndexAddOption uint

const (
	IndexAddDefault              IndexAddOption = C.GIT_INDEX_ADD_DEFAULT
	IndexAddForce                IndexAddOption = C.GIT_INDEX_ADD_FORCE
	IndexAddDisablePathspecMatch IndexAddOption = C.GIT_INDEX_ADD_DISABLE_PATHSPEC_MATCH
	IndexAddCheckPathspec        IndexAddOption = C.GIT_INDEX_ADD_CHECK_PATHSPEC
)

// IndexStageState indicates the state of the git index.
type IndexStageState int

const (
	// IndexStageAny matches any index stage.
	//
	// Some index APIs take a stage to match; pass this value to match
	// any entry matching the path regardless of stage.
	IndexStageAny IndexStageState = C.GIT_INDEX_STAGE_ANY
	// IndexStageNormal is a normal staged file in the index.
	IndexStageNormal IndexStageState = C.GIT_INDEX_STAGE_NORMAL
	// IndexStageAncestor is the ancestor side of a conflict.
	IndexStageAncestor IndexStageState = C.GIT_INDEX_STAGE_ANCESTOR
	// IndexStageOurs is the "ours" side of a conflict.
	IndexStageOurs IndexStageState = C.GIT_INDEX_STAGE_OURS
	// IndexStageTheirs is the "theirs" side of a conflict.
	IndexStageTheirs IndexStageState = C.GIT_INDEX_STAGE_THEIRS
)

type Index struct {
	doNotCompare
	ptr  *C.git_index
	repo *Repository
}

type IndexTime struct {
	Seconds     int32
	Nanoseconds uint32
}

type IndexEntry struct {
	Ctime IndexTime
	Mtime IndexTime
	Mode  Filemode
	Uid   uint32
	Gid   uint32
	Size  uint32
	Id    *Oid
	Path  string
}

func newIndexEntryFromC(entry *C.git_index_entry) *IndexEntry {
	if entry == nil {
		return nil
	}
	return &IndexEntry{
		IndexTime{int32(entry.ctime.seconds), uint32(entry.ctime.nanoseconds)},
		IndexTime{int32(entry.mtime.seconds), uint32(entry.mtime.nanoseconds)},
		Filemode(entry.mode),
		uint32(entry.uid),
		uint32(entry.gid),
		uint32(entry.file_size),
		newOidFromC(&entry.id),
		C.GoString(entry.path),
	}
}

func populateCIndexEntry(source *IndexEntry, dest *C.git_index_entry) {
	dest.ctime.seconds = C.int32_t(source.Ctime.Seconds)
	dest.ctime.nanoseconds = C.uint32_t(source.Ctime.Nanoseconds)
	dest.mtime.seconds = C.int32_t(source.Mtime.Seconds)
	dest.mtime.nanoseconds = C.uint32_t(source.Mtime.Nanoseconds)
	dest.mode = C.uint32_t(source.Mode)
	dest.uid = C.uint32_t(source.Uid)
	dest.gid = C.uint32_t(source.Gid)
	dest.file_size = C.uint32_t(source.Size)
	if source.Id != nil {
		dest.id = *source.Id.toC()
	}
	dest.path = C.CString(source.Path)
}

func freeCIndexEntry(entry *C.git_index_entry) {
	C.free(unsafe.Pointer(entry.path))
}

func newIndexFromC(ptr *C.git_index, repo *Repository) *Index {
	idx := &Index{ptr: ptr, repo: repo}
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

	return newIndexFromC(ptr, nil), nil
}

// OpenIndex creates a new index at the given path. If the file does
// not exist it will be created when Write() is called.
func OpenIndex(path string) (*Index, error) {
	var ptr *C.git_index

	var cpath = C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.git_index_open(&ptr, cpath); err < 0 {
		return nil, MakeGitError(err)
	}

	return newIndexFromC(ptr, nil), nil
}

// Path returns the index' path on disk or an empty string if it
// exists only in memory.
func (v *Index) Path() string {
	ret := C.GoString(C.git_index_path(v.ptr))
	runtime.KeepAlive(v)
	return ret
}

// Clear clears the index object in memory; changes must be explicitly
// written to disk for them to take effect persistently
func (v *Index) Clear() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C.git_index_clear(v.ptr)
	runtime.KeepAlive(v)
	if err < 0 {
		return MakeGitError(err)
	}
	return nil
}

// Add adds or replaces the given entry to the index, making a copy of
// the data
func (v *Index) Add(entry *IndexEntry) error {
	var centry C.git_index_entry

	populateCIndexEntry(entry, &centry)
	defer freeCIndexEntry(&centry)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C.git_index_add(v.ptr, &centry)
	runtime.KeepAlive(v)
	if err < 0 {
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
	runtime.KeepAlive(v)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

// AddFromBuffer adds or replaces an index entry from a buffer in memory
func (v *Index) AddFromBuffer(entry *IndexEntry, buffer []byte) error {
	var centry C.git_index_entry

	populateCIndexEntry(entry, &centry)
	defer freeCIndexEntry(&centry)

	var cbuffer unsafe.Pointer
	if len(buffer) > 0 {
		cbuffer = unsafe.Pointer(&buffer[0])
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.git_index_add_from_buffer(v.ptr, &centry, cbuffer, C.size_t(len(buffer))); err < 0 {
		return MakeGitError(err)
	}

	return nil
}

func (v *Index) AddAll(pathspecs []string, flags IndexAddOption, callback IndexMatchedPathCallback) error {
	cpathspecs := C.git_strarray{}
	cpathspecs.count = C.size_t(len(pathspecs))
	cpathspecs.strings = makeCStringsFromStrings(pathspecs)
	defer freeStrarray(&cpathspecs)

	var err error
	data := indexMatchedPathCallbackData{
		callback:    callback,
		errorTarget: &err,
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var handle unsafe.Pointer
	if callback != nil {
		handle = pointerHandles.Track(&data)
		defer pointerHandles.Untrack(handle)
	}

	ret := C._go_git_index_add_all(
		v.ptr,
		&cpathspecs,
		C.uint(flags),
		handle,
	)
	runtime.KeepAlive(v)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
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

	var err error
	data := indexMatchedPathCallbackData{
		callback:    callback,
		errorTarget: &err,
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var handle unsafe.Pointer
	if callback != nil {
		handle = pointerHandles.Track(&data)
		defer pointerHandles.Untrack(handle)
	}

	ret := C._go_git_index_update_all(
		v.ptr,
		&cpathspecs,
		handle,
	)
	runtime.KeepAlive(v)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
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

	var err error
	data := indexMatchedPathCallbackData{
		callback:    callback,
		errorTarget: &err,
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var handle unsafe.Pointer
	if callback != nil {
		handle = pointerHandles.Track(&data)
		defer pointerHandles.Untrack(handle)
	}

	ret := C._go_git_index_remove_all(
		v.ptr,
		&cpathspecs,
		handle,
	)
	runtime.KeepAlive(v)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

//export indexMatchedPathCallback
func indexMatchedPathCallback(cPath, cMatchedPathspec *C.char, payload unsafe.Pointer) C.int {
	data, ok := pointerHandles.Get(payload).(*indexMatchedPathCallbackData)
	if !ok {
		panic("invalid matched path callback")
	}

	err := data.callback(C.GoString(cPath), C.GoString(cMatchedPathspec))
	if err != nil {
		*data.errorTarget = err
		return C.int(ErrorCodeUser)
	}

	return C.int(ErrorCodeOK)
}

func (v *Index) RemoveByPath(path string) error {
	cstr := C.CString(path)
	defer C.free(unsafe.Pointer(cstr))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_index_remove_bypath(v.ptr, cstr)
	runtime.KeepAlive(v)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

// RemoveDirectory removes all entries from the index under a given directory.
func (v *Index) RemoveDirectory(dir string, stage int) error {
	cstr := C.CString(dir)
	defer C.free(unsafe.Pointer(cstr))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_index_remove_directory(v.ptr, cstr, C.int(stage))
	runtime.KeepAlive(v)
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
	runtime.KeepAlive(v)
	runtime.KeepAlive(repo)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return oid, nil
}

// ReadTree replaces the contents of the index with those of the given
// tree
func (v *Index) ReadTree(tree *Tree) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_index_read_tree(v.ptr, tree.cast_ptr)
	runtime.KeepAlive(v)
	runtime.KeepAlive(tree)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

func (v *Index) WriteTree() (*Oid, error) {
	oid := new(Oid)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_index_write_tree(oid.toC(), v.ptr)
	runtime.KeepAlive(v)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return oid, nil
}

func (v *Index) Write() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_index_write(v.ptr)
	runtime.KeepAlive(v)
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
	ret := uint(C.git_index_entrycount(v.ptr))
	runtime.KeepAlive(v)
	return ret
}

func (v *Index) EntryByIndex(index uint) (*IndexEntry, error) {
	centry := C.git_index_get_byindex(v.ptr, C.size_t(index))
	if centry == nil {
		return nil, fmt.Errorf("Index out of Bounds")
	}
	ret := newIndexEntryFromC(centry)
	runtime.KeepAlive(v)
	return ret, nil
}

func (v *Index) EntryByPath(path string, stage int) (*IndexEntry, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	centry := C.git_index_get_bypath(v.ptr, cpath, C.int(stage))
	if centry == nil {
		return nil, MakeGitError(C.int(ErrorCodeNotFound))
	}
	ret := newIndexEntryFromC(centry)
	runtime.KeepAlive(v)
	return ret, nil
}

func (v *Index) Find(path string) (uint, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var pos C.size_t
	ret := C.git_index_find(&pos, v.ptr, cpath)
	runtime.KeepAlive(v)
	if ret < 0 {
		return uint(0), MakeGitError(ret)
	}
	return uint(pos), nil
}

func (v *Index) FindPrefix(prefix string) (uint, error) {
	cprefix := C.CString(prefix)
	defer C.free(unsafe.Pointer(cprefix))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var pos C.size_t
	ret := C.git_index_find_prefix(&pos, v.ptr, cprefix)
	runtime.KeepAlive(v)
	if ret < 0 {
		return uint(0), MakeGitError(ret)
	}
	return uint(pos), nil
}

func (v *Index) HasConflicts() bool {
	ret := C.git_index_has_conflicts(v.ptr) != 0
	runtime.KeepAlive(v)
	return ret
}

// FIXME: this might return an error
func (v *Index) CleanupConflicts() {
	C.git_index_conflict_cleanup(v.ptr)
	runtime.KeepAlive(v)
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
	runtime.KeepAlive(v)
	runtime.KeepAlive(ancestor)
	runtime.KeepAlive(our)
	runtime.KeepAlive(their)
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

func (v *Index) Conflict(path string) (IndexConflict, error) {

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
	ret := IndexConflict{
		Ancestor: newIndexEntryFromC(cancestor),
		Our:      newIndexEntryFromC(cour),
		Their:    newIndexEntryFromC(ctheir),
	}
	runtime.KeepAlive(v)
	return ret, nil
}

// deprecated: You should use `Index.Conflict()` instead.
func (v *Index) GetConflict(path string) (IndexConflict, error) {
	return v.Conflict(path)
}

func (v *Index) RemoveConflict(path string) error {

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_index_conflict_remove(v.ptr, cpath)
	runtime.KeepAlive(v)
	if ecode < 0 {
		return MakeGitError(ecode)
	}
	return nil
}

type IndexConflictIterator struct {
	doNotCompare
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
	ret := IndexConflict{
		Ancestor: newIndexEntryFromC(cancestor),
		Our:      newIndexEntryFromC(cour),
		Their:    newIndexEntryFromC(ctheir),
	}
	runtime.KeepAlive(v)
	return ret, nil
}
