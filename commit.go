package git

/*
#include <git2.h>

extern int _go_git_treewalk(git_tree *tree, git_treewalk_mode mode, void *ptr);
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// Commit
type Commit struct {
	Object
	cast_ptr *C.git_commit
}

func (c *Commit) AsObject() *Object {
	return &c.Object
}

func (c *Commit) Message() string {
	ret := C.GoString(C.git_commit_message(c.cast_ptr))
	runtime.KeepAlive(c)
	return ret
}

func (c *Commit) RawMessage() string {
	ret := C.GoString(C.git_commit_message_raw(c.cast_ptr))
	runtime.KeepAlive(c)
	return ret
}

func (c *Commit) ExtractSignature() (string, string, error) {

	var c_signed C.git_buf
	defer C.git_buf_dispose(&c_signed)

	var c_signature C.git_buf
	defer C.git_buf_dispose(&c_signature)

	oid := c.Id()
	repo := C.git_commit_owner(c.cast_ptr)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ret := C.git_commit_extract_signature(&c_signature, &c_signed, repo, oid.toC(), nil)
	runtime.KeepAlive(oid)
	if ret < 0 {
		return "", "", MakeGitError(ret)
	} else {
		return C.GoString(c_signature.ptr), C.GoString(c_signed.ptr), nil
	}

}

func (c *Commit) Summary() string {
	ret := C.GoString(C.git_commit_summary(c.cast_ptr))
	runtime.KeepAlive(c)
	return ret
}

func (c *Commit) Tree() (*Tree, error) {
	var ptr *C.git_tree

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C.git_commit_tree(&ptr, c.cast_ptr)
	runtime.KeepAlive(c)
	if err < 0 {
		return nil, MakeGitError(err)
	}

	return allocTree(ptr, c.repo), nil
}

func (c *Commit) TreeId() *Oid {
	ret := newOidFromC(C.git_commit_tree_id(c.cast_ptr))
	runtime.KeepAlive(c)
	return ret
}

func (c *Commit) Author() *Signature {
	cast_ptr := C.git_commit_author(c.cast_ptr)
	ret := newSignatureFromC(cast_ptr)
	runtime.KeepAlive(c)
	return ret
}

func (c *Commit) Committer() *Signature {
	cast_ptr := C.git_commit_committer(c.cast_ptr)
	ret := newSignatureFromC(cast_ptr)
	runtime.KeepAlive(c)
	return ret
}

func (c *Commit) Parent(n uint) *Commit {
	var cobj *C.git_commit
	ret := C.git_commit_parent(&cobj, c.cast_ptr, C.uint(n))
	if ret != 0 {
		return nil
	}

	parent := allocCommit(cobj, c.repo)
	runtime.KeepAlive(c)
	return parent
}

func (c *Commit) ParentId(n uint) *Oid {
	ret := newOidFromC(C.git_commit_parent_id(c.cast_ptr, C.uint(n)))
	runtime.KeepAlive(c)
	return ret
}

func (c *Commit) ParentCount() uint {
	ret := uint(C.git_commit_parentcount(c.cast_ptr))
	runtime.KeepAlive(c)
	return ret
}

func (c *Commit) Amend(refname string, author, committer *Signature, message string, tree *Tree) (*Oid, error) {
	var cref *C.char
	if refname == "" {
		cref = nil
	} else {
		cref = C.CString(refname)
		defer C.free(unsafe.Pointer(cref))
	}

	cmsg := C.CString(message)
	defer C.free(unsafe.Pointer(cmsg))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	authorSig, err := author.toC()
	if err != nil {
		return nil, err
	}
	defer C.git_signature_free(authorSig)

	committerSig, err := committer.toC()
	if err != nil {
		return nil, err
	}
	defer C.git_signature_free(committerSig)

	oid := new(Oid)

	cerr := C.git_commit_amend(oid.toC(), c.cast_ptr, cref, authorSig, committerSig, nil, cmsg, tree.cast_ptr)
	runtime.KeepAlive(oid)
	runtime.KeepAlive(c)
	runtime.KeepAlive(tree)
	if cerr < 0 {
		return nil, MakeGitError(cerr)
	}

	return oid, nil
}
