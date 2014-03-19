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

type branchIterator struct {
	ptr *C.git_branch_iterator
}

func newBranchIteratorFromC(ptr *C.git_branch_iterator) ReferenceIterator {
	i := &branchIterator{ptr: ptr}
	runtime.SetFinalizer(i, (*branchIterator).Free)
	return i
}

func (i *branchIterator) NextName() (string, error) {
	ref, err := i.Next()
	if err != nil {
		return "", err
	}
	defer ref.Free()

	return ref.Name(), err
}

func (i *branchIterator) Next() (*Reference, error) {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var refPtr *C.git_reference
	var refType C.git_branch_t

	ecode := C.git_branch_next(&refPtr, &refType, i.ptr)

	if ecode == C.GIT_ITEROVER {
		return nil, ErrIterOver
	} else if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newReferenceFromC(refPtr), nil
}

func (i *branchIterator) Free() {
	runtime.SetFinalizer(i, nil)
	C.git_branch_iterator_free(i.ptr)
}

func (repo *Repository) NewBranchIterator(flags BranchType) (ReferenceIterator, error) {

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

func (b *Reference) DeleteBranch() error {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ret := C.git_branch_delete(b.ptr)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (b *Reference) MoveBranch(newBranchName string, force bool, signature *Signature, msg string) (*Reference, error) {
	var ptr *C.git_reference
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

	ret := C.git_branch_move(&ptr, b.ptr, cNewBranchName, cForce, cSignature, cmsg)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return newReferenceFromC(ptr), nil
}

func (b *Reference) IsBranchHead() (bool, error) {

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

func (repo *Repository) LookupBranch(branchName string, bt BranchType) (*Reference, error) {
	var ptr *C.git_reference

	cName := C.CString(branchName)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_branch_lookup(&ptr, repo.ptr, cName, C.git_branch_t(bt))
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return newReferenceFromC(ptr), nil
}

func (b *Reference) BranchName() (string, error) {
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

func (b *Reference) SetUpstream(upstreamName string) error {
	cName := C.CString(upstreamName)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_branch_set_upstream(b.ptr, cName)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (b *Reference) Upstream() (*Reference, error) {

	var ptr *C.git_reference
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_branch_upstream(&ptr, b.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return newReferenceFromC(ptr), nil
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
