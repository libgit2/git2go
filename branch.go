package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
*/
import "C"

import (
	"unsafe"
)

type BranchType uint

const (
	BRANCH_LOCAL  BranchType = C.GIT_BRANCH_LOCAL
	BRANCH_REMOTE            = C.GIT_BRANCH_REMOTE
)

const (
	REFS_DIR         = "refs/"
	REFS_HEADS_DIR   = REFS_DIR + "heads/"
	REFS_TAGS_DIR    = REFS_DIR + "tags/"
	REFS_REMOTES_DIR = REFS_DIR + "remotes/"
)

type Branch struct {
	Reference
}

func (repo *Repository) CreateBranch(branchName string, target *Commit, force bool) (*Reference, error) {
	ref := new(Reference)
	cBranchName := C.CString(branchName)
	cForce := cbool(force)
	err := C.git_branch_create(&ref.ptr, repo.ptr, cBranchName, target.ptr, cForce)
	if err < 0 {
		return nil, LastError()
	}
	return ref, nil
}

func (b *Branch) BranchDelete() error {
	if err := C.git_branch_delete(b.ptr); err < 0 {
		return LastError()
	}
	return nil
}

func (b *Branch) Move(newBranchName string, force bool) (*Branch, error) {
	newBranch := new(Branch)
	cNewBranchName := C.CString(newBranchName)
	cForce := cbool(force)

	err := C.git_branch_move(&newBranch.ptr, b.ptr, cNewBranchName, cForce)
	if err < 0 {
		return nil, LastError()
	}
	return newBranch, nil
}

func (b *Branch) IsHead() (bool, error) {
	isHead := C.git_branch_is_head(b.ptr)
	switch isHead {
	case 1:
		return true, nil
	case 0:
		return false, nil
	default:
		return false, LastError()
	}

}

func (repo *Repository) LookupBranch(branchName string, bt BranchType) (*Branch, error) {
	branch := new(Branch)
	cName := C.CString(branchName)

	err := C.git_branch_lookup(&branch.ptr, repo.ptr, cName, C.git_branch_t(bt))
	if err < 0 {
		return nil, LastError()
	}
	return branch, nil
}

func (b *Branch) Name() (string, error) {
	var cName *C.char
	defer C.free(unsafe.Pointer(cName))

	err := C.git_branch_name(&cName, b.ptr)
	if err < 0 {
		return "", LastError()
	}

	return C.GoString(cName), nil
}

func (repo *Repository) RemoteName(canonicalBranchName string) (string, error) {
	cName := C.CString(canonicalBranchName)

	// Obtain the length of the name
	ret := C.git_branch_remote_name(nil, 0, repo.ptr, cName)
	if ret < 0 {
		return "", LastError()
	}

	cBuf := (*C.char)(C.malloc(C.size_t(ret)))
	defer C.free(unsafe.Pointer(cBuf))

	// Actually obtain the name
	ret = C.git_branch_remote_name(cBuf, C.size_t(ret), repo.ptr, cName)
	if ret < 0 {
		return "", LastError()
	}

	return C.GoString(cBuf), nil
}

func (b *Branch) SetUpstream(upstreamName string) error {
	cName := C.CString(upstreamName)

	err := C.git_branch_set_upstream(b.ptr, cName)
	if err < 0 {
		return LastError()
	}
	return nil
}

func (b *Branch) Upstream() (*Branch, error) {
	upstream := new(Branch)
	err := C.git_branch_upstream(&upstream.ptr, b.ptr)
	if err < 0 {
		return nil, LastError()
	}
	return upstream, nil
}

func (repo *Repository) UpstreamName(canonicalBranchName string) (string, error) {
	cName := C.CString(canonicalBranchName)

	// Obtain the length of the name
	ret := C.git_branch_upstream_name(nil, 0, repo.ptr, cName)
	if ret < 0 {
		return "", LastError()
	}

	cBuf := (*C.char)(C.malloc(C.size_t(ret)))
	defer C.free(unsafe.Pointer(cBuf))

	// Actually obtain the name
	ret = C.git_branch_upstream_name(cBuf, C.size_t(ret), repo.ptr, cName)
	if ret < 0 {
		return "", LastError()
	}
	return C.GoString(cBuf), nil
}
