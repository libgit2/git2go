package git

/*
#include <git2.h>
#include <git2/errors.h>

extern int _go_git_treewalk(git_tree *tree, git_treewalk_mode mode, void *ptr);
*/
import "C"

import (
	"runtime"
	"time"
	"unsafe"
)

// Commit
type Commit struct {
	gitObject
	cast_ptr *C.git_commit
}

func (c Commit) Message() string {
	return C.GoString(C.git_commit_message(c.cast_ptr))
}

func (c Commit) Tree() (*Tree, error) {
	var ptr *C.git_tree

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C.git_commit_tree(&ptr, c.cast_ptr)
	if err < 0 {
		return nil, MakeGitError(err)
	}

	return allocObject((*C.git_object)(ptr), c.repo).(*Tree), nil
}

func (c Commit) TreeId() *Oid {
	return newOidFromC(C.git_commit_tree_id(c.cast_ptr))
}

func (c Commit) Author() *Signature {
	cast_ptr := C.git_commit_author(c.cast_ptr)
	return newSignatureFromC(cast_ptr)
}

func (c Commit) Committer() *Signature {
	cast_ptr := C.git_commit_committer(c.cast_ptr)
	return newSignatureFromC(cast_ptr)
}

func (c *Commit) Parent(n uint) *Commit {
	var cobj *C.git_commit
	ret := C.git_commit_parent(&cobj, c.cast_ptr, C.uint(n))
	if ret != 0 {
		return nil
	}

	return allocObject((*C.git_object)(cobj), c.repo).(*Commit)
}

func (c *Commit) ParentId(n uint) *Oid {
	return newOidFromC(C.git_commit_parent_id(c.cast_ptr, C.uint(n)))
}

func (c *Commit) ParentCount() uint {
	return uint(C.git_commit_parentcount(c.cast_ptr))
}

// Signature

type Signature struct {
	Name  string
	Email string
	When  time.Time
}

func newSignatureFromC(sig *C.git_signature) *Signature {
	// git stores minutes, go wants seconds
	loc := time.FixedZone("", int(sig.when.offset)*60)
	return &Signature{
		C.GoString(sig.name),
		C.GoString(sig.email),
		time.Unix(int64(sig.when.time), 0).In(loc),
	}
}

// the offset in mintes, which is what git wants
func (v *Signature) Offset() int {
	_, offset := v.When.Zone()
	return offset / 60
}

func (sig *Signature) toC() *C.git_signature {

	if sig == nil {
		return nil
	}

	var out *C.git_signature

	name := C.CString(sig.Name)
	defer C.free(unsafe.Pointer(name))

	email := C.CString(sig.Email)
	defer C.free(unsafe.Pointer(email))

	ret := C.git_signature_new(&out, name, email, C.git_time_t(sig.When.Unix()), C.int(sig.Offset()))
	if ret < 0 {
		return nil
	}

	return out
}
