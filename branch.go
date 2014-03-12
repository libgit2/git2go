package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

type BranchType uint

const (
	BranchLocal  BranchType = C.GIT_BRANCH_LOCAL
	BranchRemote            = C.GIT_BRANCH_REMOTE
)

type Branch struct {
	Reference
}

type BranchIterator struct {
	ptr *C.git_branch_iterator
}

func newBranchIteratorFromC(ptr *C.git_branch_iterator) *BranchIterator {
	i := &BranchIterator{ptr: ptr}
	runtime.SetFinalizer(i, (*BranchIterator).Free)
	return i
}

func (i *BranchIterator) Next() (*Reference, error) {
	ref, _, err := i.NextWithType()
	return ref, err
}

func (i *BranchIterator) NextWithType() (*Reference, BranchType, error) {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var refPtr *C.git_reference
	var refType C.git_branch_t

	ecode := C.git_branch_next(&refPtr, &refType, i.ptr)

	if ecode == C.GIT_ITEROVER {
		return nil, BranchLocal, ErrIterOver
	} else if ecode < 0 {
		return nil, BranchLocal, MakeGitError(ecode)
	}

	return newReferenceFromC(refPtr), BranchType(refType), nil
}

func (i *BranchIterator) Free() {
	runtime.SetFinalizer(i, nil)
	C.git_branch_iterator_free(i.ptr)
}

func (repo *Repository) NewBranchIterator(flags BranchType) (*BranchIterator, error) {

	refType := C.git_branch_t(flags)
	var ptr *C.git_branch_iterator

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_branch_iterator_new(&ptr, repo.ptr, refType)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newBranchIteratorFromC(ptr), nil
}

func (repo *Repository) CreateBranch(branchName string, target *Commit, force bool, signature *Signature, msg string) (*Reference, error) {

	ref := new(Reference)
	cBranchName := C.CString(branchName)
	cForce := cbool(force)

	cSignature := signature.toC()
	defer C.git_signature_free(cSignature)

	var cmsg *C.char
	if msg == "" {
		cmsg = nil
	} else {
		cmsg = C.CString(msg)
		defer C.free(unsafe.Pointer(cmsg))
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_branch_create(&ref.ptr, repo.ptr, cBranchName, target.ptr, cForce, cSignature, cmsg)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return ref, nil
}

func (b *Branch) Delete() error {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ret := C.git_branch_delete(b.ptr)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (b *Branch) Move(newBranchName string, force bool, signature *Signature, msg string) (*Branch, error) {
	newBranch := new(Branch)
	cNewBranchName := C.CString(newBranchName)
	cForce := cbool(force)

	cSignature := signature.toC()
	defer C.git_signature_free(cSignature)

	var cmsg *C.char
	if msg == "" {
		cmsg = nil
	} else {
		cmsg = C.CString(msg)
		defer C.free(unsafe.Pointer(cmsg))
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_branch_move(&newBranch.ptr, b.ptr, cNewBranchName, cForce, cSignature, cmsg)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return newBranch, nil
}

func (b *Branch) IsHead() (bool, error) {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_branch_is_head(b.ptr)
	switch ret {
	case 1:
		return true, nil
	case 0:
		return false, nil
	}
	return false, MakeGitError(ret)

}

func (repo *Repository) LookupBranch(branchName string, bt BranchType) (*Branch, error) {
	branch := new(Branch)
	cName := C.CString(branchName)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_branch_lookup(&branch.ptr, repo.ptr, cName, C.git_branch_t(bt))
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return branch, nil
}

func (b *Branch) Name() (string, error) {
	var cName *C.char
	defer C.free(unsafe.Pointer(cName))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_branch_name(&cName, b.ptr)
	if ret < 0 {
		return "", MakeGitError(ret)
	}

	return C.GoString(cName), nil
}

func (repo *Repository) RemoteName(canonicalBranchName string) (string, error) {
	cName := C.CString(canonicalBranchName)

	nameBuf := C.git_buf{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_branch_remote_name(&nameBuf, repo.ptr, cName)
	if ret < 0 {
		return "", MakeGitError(ret)
	}
	defer C.git_buf_free(&nameBuf)

	return C.GoString(nameBuf.ptr), nil
}

func (b *Branch) SetUpstream(upstreamName string) error {
	cName := C.CString(upstreamName)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_branch_set_upstream(b.ptr, cName)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (b *Branch) Upstream() (*Branch, error) {
	upstream := new(Branch)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_branch_upstream(&upstream.ptr, b.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return upstream, nil
}

func (repo *Repository) UpstreamName(canonicalBranchName string) (string, error) {
	cName := C.CString(canonicalBranchName)

	nameBuf := C.git_buf{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_branch_upstream_name(&nameBuf, repo.ptr, cName)
	if ret < 0 {
		return "", MakeGitError(ret)
	}
	defer C.git_buf_free(&nameBuf)

	return C.GoString(nameBuf.ptr), nil
}
