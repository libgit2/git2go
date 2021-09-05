package git

/*
#include <git2.h>

extern int _go_git_treewalk(git_tree *tree, git_treewalk_mode mode, void *ptr);
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

type Filemode int

const (
	FilemodeTree           Filemode = C.GIT_FILEMODE_TREE
	FilemodeBlob           Filemode = C.GIT_FILEMODE_BLOB
	FilemodeBlobExecutable Filemode = C.GIT_FILEMODE_BLOB_EXECUTABLE
	FilemodeLink           Filemode = C.GIT_FILEMODE_LINK
	FilemodeCommit         Filemode = C.GIT_FILEMODE_COMMIT
)

type Tree struct {
	doNotCompare
	Object
	cast_ptr *C.git_tree
}

func (t *Tree) AsObject() *Object {
	return &t.Object
}

type TreeEntry struct {
	Name     string
	Id       *Oid
	Type     ObjectType
	Filemode Filemode
}

func newTreeEntry(entry *C.git_tree_entry) *TreeEntry {
	return &TreeEntry{
		C.GoString(C.git_tree_entry_name(entry)),
		newOidFromC(C.git_tree_entry_id(entry)),
		ObjectType(C.git_tree_entry_type(entry)),
		Filemode(C.git_tree_entry_filemode(entry)),
	}
}

func (t *Tree) EntryByName(filename string) *TreeEntry {
	cname := C.CString(filename)
	defer C.free(unsafe.Pointer(cname))

	entry := C.git_tree_entry_byname(t.cast_ptr, cname)
	if entry == nil {
		return nil
	}

	goEntry := newTreeEntry(entry)
	runtime.KeepAlive(t)
	return goEntry
}

// EntryById performs a lookup for a tree entry with the given SHA value.
//
// It returns a *TreeEntry that is owned by the Tree. You don't have to
// free it, but you must not use it after the Tree is freed.
//
// Warning: this must examine every entry in the tree, so it is not fast.
func (t *Tree) EntryById(id *Oid) *TreeEntry {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	entry := C.git_tree_entry_byid(t.cast_ptr, id.toC())
	runtime.KeepAlive(id)
	if entry == nil {
		return nil
	}

	goEntry := newTreeEntry(entry)
	runtime.KeepAlive(t)
	return goEntry
}

// EntryByPath looks up an entry by its full path, recursing into
// deeper trees if necessary (i.e. if there are slashes in the path)
func (t *Tree) EntryByPath(path string) (*TreeEntry, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	var entry *C.git_tree_entry

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_tree_entry_bypath(&entry, t.cast_ptr, cpath)
	runtime.KeepAlive(t)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	defer C.git_tree_entry_free(entry)

	return newTreeEntry(entry), nil
}

func (t *Tree) EntryByIndex(index uint64) *TreeEntry {
	entry := C.git_tree_entry_byindex(t.cast_ptr, C.size_t(index))
	if entry == nil {
		return nil
	}

	goEntry := newTreeEntry(entry)
	runtime.KeepAlive(t)
	return goEntry
}

func (t *Tree) EntryCount() uint64 {
	num := C.git_tree_entrycount(t.cast_ptr)
	runtime.KeepAlive(t)
	return uint64(num)
}

type TreeWalkCallback func(string, *TreeEntry) int
type treeWalkCallbackData struct {
	callback    TreeWalkCallback
	errorTarget *error
}

//export treeWalkCallback
func treeWalkCallback(_root *C.char, entry *C.git_tree_entry, ptr unsafe.Pointer) C.int {
	data, ok := pointerHandles.Get(ptr).(*treeWalkCallbackData)
	if !ok {
		panic("invalid treewalk callback")
	}

	ret := data.callback(C.GoString(_root), newTreeEntry(entry))
	if ret < 0 {
		*data.errorTarget = errors.New(ErrorCode(ret).String())
		return C.int(ErrorCodeUser)
	}

	return C.int(ErrorCodeOK)
}

func (t *Tree) Walk(callback TreeWalkCallback) error {
	var err error
	data := treeWalkCallbackData{
		callback:    callback,
		errorTarget: &err,
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	handle := pointerHandles.Track(&data)
	defer pointerHandles.Untrack(handle)

	ret := C._go_git_treewalk(t.cast_ptr, C.GIT_TREEWALK_PRE, handle)
	runtime.KeepAlive(t)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

type TreeBuilder struct {
	doNotCompare
	ptr  *C.git_treebuilder
	repo *Repository
}

func (v *TreeBuilder) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_treebuilder_free(v.ptr)
}

func (v *TreeBuilder) Insert(filename string, id *Oid, filemode Filemode) error {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C.git_treebuilder_insert(nil, v.ptr, cfilename, id.toC(), C.git_filemode_t(filemode))
	runtime.KeepAlive(v)
	runtime.KeepAlive(id)
	if err < 0 {
		return MakeGitError(err)
	}

	return nil
}

func (v *TreeBuilder) Remove(filename string) error {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C.git_treebuilder_remove(v.ptr, cfilename)
	runtime.KeepAlive(v)
	if err < 0 {
		return MakeGitError(err)
	}

	return nil
}

func (v *TreeBuilder) Write() (*Oid, error) {
	oid := new(Oid)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C.git_treebuilder_write(oid.toC(), v.ptr)
	runtime.KeepAlive(v)
	if err < 0 {
		return nil, MakeGitError(err)
	}

	return oid, nil
}
