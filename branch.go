package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
*/
import "C"

import (
	"errors"
	"strings"
	"unsafe"
)

var ErrEUser = errors.New("Error in user callback function")

type ListFlags uint

type BranchT uint

const (
	BRANCH_LOCAL  BranchT = C.GIT_BRANCH_LOCAL
	BRANCH_REMOTE         = C.GIT_BRANCH_REMOTE
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

func (repo *Repository) BranchCreate(branchName string, target *Commit, force bool) (*Reference, error) {
	ref := new(Reference)
	cBranchName := C.CString(branchName)
	cForce := cbool(force)
	err := C.git_branch_create(&ref.ptr, repo.ptr, cBranchName, target.ptr, cForce)
	if err < 0 {
		return nil, LastError()
	}
	return ref, nil
}

func (branch *Branch) BranchDelete() error {
	if err := C.git_branch_delete(branch.ptr); err < 0 {
		return LastError()
	}
	return nil
}

type BranchForeachCB func(name string, flags ListFlags, payload interface{}) error

func (repo *Repository) BranchForeach(flags ListFlags, callback BranchForeachCB, payload interface{}) error {
	iter, err := repo.NewReferenceIterator()
	if err != nil {
		return err
	}

	for {
		ref, err := iter.Next()
		if err == ErrIterOver {
			break
		}

		if (flags == ListFlags(BRANCH_LOCAL)) && strings.HasPrefix(ref.Name(), REFS_HEADS_DIR) {
			name := strings.TrimPrefix(ref.Name(), REFS_HEADS_DIR)
			err = callback(name, ListFlags(BRANCH_LOCAL), payload)
			if err != nil {
				return err
			}
		}

		if (flags == ListFlags(BRANCH_REMOTE)) && strings.HasPrefix(ref.Name(), REFS_REMOTES_DIR) {
			name := strings.TrimPrefix(ref.Name(), REFS_REMOTES_DIR)
			err = callback(name, ListFlags(BRANCH_REMOTE), payload)
			if err != nil {
				return err
			}
		}
	}

	if err == ErrIterOver {
		err = nil
	}
	return err
}

func (branch *Branch) Move(newBranchName string, force bool) (*Branch, error) {
	newBranch := new(Branch)
	cNewBranchName := C.CString(newBranchName)
	cForce := cbool(force)

	err := C.git_branch_move(&newBranch.ptr, branch.ptr, cNewBranchName, cForce)
	if err < 0 {
		return nil, LastError()
	}
	return newBranch, nil
}

func (branch *Branch) IsHead() (bool, error) {
	isHead := C.git_branch_is_head(branch.ptr)
	switch isHead {
	case 1:
		return true, nil
	case 0:
		return false, nil
	default:
		return false, LastError()
	}

}

func (repo *Repository) BranchLookup(branchName string, branchType BranchT) (*Branch, error) {
	branch := new(Branch)
	cName := C.CString(branchName)

	err := C.git_branch_lookup(&branch.ptr, repo.ptr, cName, C.git_branch_t(branchType))
	if err < 0 {
		return nil, LastError()
	}
	return branch, nil
}

func (branch *Branch) Name() (string, error) {
	var cName *C.char
	defer C.free(unsafe.Pointer(cName))

	err := C.git_branch_name(&cName, branch.ptr)
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

func (branch *Branch) SetUpstream(upstreamName string) error {
	cName := C.CString(upstreamName)

	err := C.git_branch_set_upstream(branch.ptr, cName)
	if err < 0 {
		return LastError()
	}
	return nil
}

func (branch *Branch) Upstream() (*Branch, error) {
	upstream := new(Branch)
	err := C.git_branch_upstream(&upstream.ptr, branch.ptr)
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
