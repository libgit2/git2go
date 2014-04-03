package git

/*
#include <git2.h>
#include <git2/errors.h>

extern int _go_git_visit_submodule(git_repository *repo, void *fct);
*/
import "C"
import (
	"runtime"
	"unsafe"
)

// Submodule
type Submodule struct {
	ptr *C.git_submodule
}

type SubmoduleUpdate int

const (
	SubmoduleUpdateReset    SubmoduleUpdate = C.GIT_SUBMODULE_UPDATE_RESET
	SubmoduleUpdateCheckout                 = C.GIT_SUBMODULE_UPDATE_CHECKOUT
	SubmoduleUpdateRebase                   = C.GIT_SUBMODULE_UPDATE_REBASE
	SubmoduleUpdateMerge                    = C.GIT_SUBMODULE_UPDATE_MERGE
	SubmoduleUpdateNone                     = C.GIT_SUBMODULE_UPDATE_NONE
)

type SubmoduleIgnore int

const (
	SubmoduleIgnoreReset     SubmoduleIgnore = C.GIT_SUBMODULE_IGNORE_RESET
	SubmoduleIgnoreNone                      = C.GIT_SUBMODULE_IGNORE_NONE
	SubmoduleIgnoreUntracked                 = C.GIT_SUBMODULE_IGNORE_UNTRACKED
	SubmoduleIgnoreDirty                     = C.GIT_SUBMODULE_IGNORE_DIRTY
	SubmoduleIgnoreAll                       = C.GIT_SUBMODULE_IGNORE_ALL
)

type SubmoduleStatus int

const (
	SubmoduleStatusInHead          SubmoduleStatus = C.GIT_SUBMODULE_STATUS_IN_HEAD
	SubmoduleStatusInIndex                         = C.GIT_SUBMODULE_STATUS_IN_INDEX
	SubmoduleStatusInConfig                        = C.GIT_SUBMODULE_STATUS_IN_CONFIG
	SubmoduleStatusInWd                            = C.GIT_SUBMODULE_STATUS_IN_WD
	SubmoduleStatusIndexAdded                      = C.GIT_SUBMODULE_STATUS_INDEX_ADDED
	SubmoduleStatusIndexDeleted                    = C.GIT_SUBMODULE_STATUS_INDEX_DELETED
	SubmoduleStatusIndexModified                   = C.GIT_SUBMODULE_STATUS_INDEX_MODIFIED
	SubmoduleStatusWdUninitialized                 = C.GIT_SUBMODULE_STATUS_WD_UNINITIALIZED
	SubmoduleStatusWdAdded                         = C.GIT_SUBMODULE_STATUS_WD_ADDED
	SubmoduleStatusWdDeleted                       = C.GIT_SUBMODULE_STATUS_WD_DELETED
	SubmoduleStatusWdModified                      = C.GIT_SUBMODULE_STATUS_WD_MODIFIED
	SubmoduleStatusWdIndexModified                 = C.GIT_SUBMODULE_STATUS_WD_INDEX_MODIFIED
	SubmoduleStatusWdWdModified                    = C.GIT_SUBMODULE_STATUS_WD_WD_MODIFIED
	SubmoduleStatusWdUntracked                     = C.GIT_SUBMODULE_STATUS_WD_UNTRACKED
)

type SubmoduleRecurse int

const (
	SubmoduleRecurseNo       SubmoduleRecurse = C.GIT_SUBMODULE_RECURSE_NO
	SubmoduleRecurseYes                       = C.GIT_SUBMODULE_RECURSE_YES
	SubmoduleRecurseOndemand                  = C.GIT_SUBMODULE_RECURSE_ONDEMAND
)

func SubmoduleStatusIsUnmodified(status int) bool {
	o := SubmoduleStatus(status) & ^(SubmoduleStatusInHead | SubmoduleStatusInIndex |
		SubmoduleStatusInConfig | SubmoduleStatusInWd)
	return o == 0
}

func (repo *Repository) LookupSubmodule(name string) (*Submodule, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	sub := new(Submodule)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_submodule_lookup(&sub.ptr, repo.ptr, cname)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return sub, nil
}

type SubmoduleCbk func(sub *Submodule, name string) int

//export SubmoduleVisitor
func SubmoduleVisitor(csub unsafe.Pointer, name string, cfct unsafe.Pointer) int {
	sub := &Submodule{(*C.git_submodule)(csub)}
	fct := *(*SubmoduleCbk)(cfct)
	return fct(sub, name)
}

func (repo *Repository) ForeachSubmodule(cbk SubmoduleCbk) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C._go_git_visit_submodule(repo.ptr, unsafe.Pointer(&cbk))
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (repo *Repository) AddSubmodule(url, path string, use_git_link bool) (*Submodule, error) {
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	sub := new(Submodule)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_submodule_add_setup(&sub.ptr, repo.ptr, curl, cpath, cbool(use_git_link))
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return sub, nil
}

func (sub *Submodule) FinalizeAdd() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_submodule_add_finalize(sub.ptr)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (sub *Submodule) AddToIndex(write_index bool) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_submodule_add_to_index(sub.ptr, cbool(write_index))
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (sub *Submodule) Save() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_submodule_save(sub.ptr)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (sub *Submodule) Owner() *Repository {
	repo := C.git_submodule_owner(sub.ptr)
	//FIXME: how to handle dangling references ?
	return &Repository{repo}
}

func (sub *Submodule) Name() string {
	n := C.git_submodule_name(sub.ptr)
	return C.GoString(n)
}

func (sub *Submodule) Path() string {
	n := C.git_submodule_path(sub.ptr)
	return C.GoString(n)
}

func (sub *Submodule) Url() string {
	n := C.git_submodule_url(sub.ptr)
	return C.GoString(n)
}

func (sub *Submodule) SetUrl(url string) error {
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_submodule_set_url(sub.ptr, curl)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (sub *Submodule) IndexId() *Oid {
	idx := C.git_submodule_index_id(sub.ptr)
	if idx == nil {
		return nil
	}
	return newOidFromC(idx)
}

func (sub *Submodule) HeadId() *Oid {
	idx := C.git_submodule_head_id(sub.ptr)
	if idx == nil {
		return nil
	}
	return newOidFromC(idx)
}

func (sub *Submodule) WdId() *Oid {
	idx := C.git_submodule_wd_id(sub.ptr)
	if idx == nil {
		return nil
	}
	return newOidFromC(idx)
}

func (sub *Submodule) Ignore() SubmoduleIgnore {
	o := C.git_submodule_ignore(sub.ptr)
	return SubmoduleIgnore(o)
}

func (sub *Submodule) SetIgnore(ignore SubmoduleIgnore) SubmoduleIgnore {
	o := C.git_submodule_set_ignore(sub.ptr, C.git_submodule_ignore_t(ignore))
	return SubmoduleIgnore(o)
}

func (sub *Submodule) Update() SubmoduleUpdate {
	o := C.git_submodule_update(sub.ptr)
	return SubmoduleUpdate(o)
}

func (sub *Submodule) SetUpdate(update SubmoduleUpdate) SubmoduleUpdate {
	o := C.git_submodule_set_update(sub.ptr, C.git_submodule_update_t(update))
	return SubmoduleUpdate(o)
}

func (sub *Submodule) FetchRecurseSubmodules() SubmoduleRecurse {
	return SubmoduleRecurse(C.git_submodule_fetch_recurse_submodules(sub.ptr))
}

func (sub *Submodule) SetFetchRecurseSubmodules(recurse SubmoduleRecurse) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_submodule_set_fetch_recurse_submodules(sub.ptr, C.git_submodule_recurse_t(recurse))
	if ret < 0 {
		return MakeGitError(C.int(ret))
	}
	return nil
}

func (sub *Submodule) Init(overwrite bool) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_submodule_init(sub.ptr, cbool(overwrite))
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (sub *Submodule) Sync() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_submodule_sync(sub.ptr)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (sub *Submodule) Open() (*Repository, error) {
	repo := new(Repository)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_submodule_open(&repo.ptr, sub.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return repo, nil
}

func (sub *Submodule) Reload(force bool) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_submodule_reload(sub.ptr, cbool(force))
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (repo *Repository) ReloadAllSubmodules(force bool) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_submodule_reload_all(repo.ptr, cbool(force))
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}
