package git

/*
#include <git2.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type WorktreeCollection struct {
	doNotCompare
	repo *Repository
}

type Worktree struct {
	doNotCompare
	ptr *C.git_worktree
}

type AddWorktreeOptions struct {
	// Lock the newly created worktree
	Lock bool
	// Reference to use for the new worktree
	Reference *Reference
	// CheckoutOptions is used for configuring the checkout for the newly created worktree
	CheckoutOptions CheckoutOptions
}

// Add adds a new working tree for the given repository
func (c *WorktreeCollection) Add(name string, path string, options *AddWorktreeOptions) (*Worktree, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var err error
	cOptions := populateAddWorktreeOptions(&C.git_worktree_add_options{}, options, &err)
	defer freeAddWorktreeOptions(cOptions)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_worktree
	ret := C.git_worktree_add(&ptr, c.repo.ptr, cName, cPath, cOptions)
	runtime.KeepAlive(c)
	if options != nil && options.Reference != nil {
		runtime.KeepAlive(options.Reference)
	}

	if ret == C.int(ErrorCodeUser) && err != nil {
		return nil, err
	} else if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return newWorktreeFromC(ptr), nil
}

// List lists names of linked working trees for the given repository
func (c *WorktreeCollection) List() ([]string, error) {
	var strC C.git_strarray
	defer C.git_strarray_dispose(&strC)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_list(&strC, c.repo.ptr)
	runtime.KeepAlive(c)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	w := makeStringsFromCStrings(strC.strings, int(strC.count))
	return w, nil
}

// Lookup gets a working tree by its name for the given repository
func (c *WorktreeCollection) Lookup(name string) (*Worktree, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_worktree
	ret := C.git_worktree_lookup(&ptr, c.repo.ptr, cname)
	runtime.KeepAlive(c)

	if ret < 0 {
		return nil, MakeGitError(ret)
	} else if ptr == nil {
		return nil, nil
	}
	return newWorktreeFromC(ptr), nil
}

// OpenFromRepository retrieves a worktree for the given repository
func (c *WorktreeCollection) OpenFromRepository() (*Worktree, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_worktree
	ret := C.git_worktree_open_from_repository(&ptr, c.repo.ptr)
	runtime.KeepAlive(c)

	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return newWorktreeFromC(ptr), nil
}

func newWorktreeFromC(ptr *C.git_worktree) *Worktree {
	worktree := &Worktree{ptr: ptr}
	runtime.SetFinalizer(worktree, (*Worktree).Free)
	return worktree
}

func freeAddWorktreeOptions(cOptions *C.git_worktree_add_options) {
	if cOptions == nil {
		return
	}
	freeCheckoutOptions(&cOptions.checkout_options)
}

func populateAddWorktreeOptions(cOptions *C.git_worktree_add_options, options *AddWorktreeOptions, errorTarget *error) *C.git_worktree_add_options {
	C.git_worktree_add_options_init(cOptions, C.GIT_WORKTREE_ADD_OPTIONS_VERSION)
	if options == nil {
		return nil
	}

	populateCheckoutOptions(&cOptions.checkout_options, &options.CheckoutOptions, errorTarget)
	cOptions.lock = cbool(options.Lock)
	if options.Reference != nil {
		cOptions.ref = options.Reference.ptr
	}
	return cOptions
}

// Free a previously allocated worktree
func (w *Worktree) Free() {
	runtime.SetFinalizer(w, nil)
	C.git_worktree_free(w.ptr)
}

// IsLocked checks if the given worktree is locked
func (w *Worktree) IsLocked() (locked bool, reason string, err error) {
	buf := C.git_buf{}
	defer C.git_buf_dispose(&buf)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_is_locked(&buf, w.ptr)
	runtime.KeepAlive(w)

	if ret < 0 {
		return false, "", MakeGitError(ret)
	}
	return ret != 0, C.GoString(buf.ptr), nil
}

type WorktreePruneFlag uint32

const (
	// WorktreePruneValid means prune working tree even if working tree is valid
	WorktreePruneValid WorktreePruneFlag = C.GIT_WORKTREE_PRUNE_VALID
	// WorktreePruneLocked means prune working tree even if it is locked
	WorktreePruneLocked WorktreePruneFlag = C.GIT_WORKTREE_PRUNE_LOCKED
	// WorktreePruneWorkingTree means prune checked out working tree
	WorktreePruneWorkingTree WorktreePruneFlag = C.GIT_WORKTREE_PRUNE_WORKING_TREE
)

// IsPrunable checks that the worktree is prunable with the given flags
func (w *Worktree) IsPrunable(flags WorktreePruneFlag) (bool, error) {
	cOptions := C.git_worktree_prune_options{}
	C.git_worktree_prune_options_init(&cOptions, C.GIT_WORKTREE_PRUNE_OPTIONS_VERSION)
	cOptions.flags = C.uint32_t(flags)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_is_prunable(w.ptr, &cOptions)
	runtime.KeepAlive(w)

	if ret < 0 {
		return false, MakeGitError(ret)
	}
	return ret != 0, nil
}

// Lock locks the worktree if not already locked
func (w *Worktree) Lock(reason string) error {
	var cReason *C.char
	if reason != "" {
		cReason = C.CString(reason)
		defer C.free(unsafe.Pointer(cReason))
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_lock(w.ptr, cReason)
	runtime.KeepAlive(w)

	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

// Name retrieves the name of the worktree
func (w *Worktree) Name() string {
	s := C.GoString(C.git_worktree_name(w.ptr))
	runtime.KeepAlive(w)
	return s
}

// Path retrieves the path of the worktree
func (w *Worktree) Path() string {
	s := C.GoString(C.git_worktree_path(w.ptr))
	runtime.KeepAlive(w)
	return s
}

// Prune the worktree with the provided flags
func (w *Worktree) Prune(flags WorktreePruneFlag) error {
	cOptions := C.git_worktree_prune_options{}
	C.git_worktree_prune_options_init(&cOptions, C.GIT_WORKTREE_PRUNE_OPTIONS_VERSION)
	cOptions.flags = C.uint32_t(flags)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_prune(w.ptr, &cOptions)
	runtime.KeepAlive(w)

	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

// Unlock a locked worktree
func (w *Worktree) Unlock() (notLocked bool, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_unlock(w.ptr)
	runtime.KeepAlive(w)

	if ret < 0 {
		return false, MakeGitError(ret)
	}
	return ret != 0, nil
}

// Validate checks if the given worktree is valid
func (w *Worktree) Validate() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_worktree_validate(w.ptr)
	runtime.KeepAlive(w)

	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}
