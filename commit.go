package git

/*
#include <git2.h>
#include <git2/errors.h>

extern int _go_git_treewalk(git_tree *tree, git_treewalk_mode mode, void *ptr);
*/
import "C"

import (
	"runtime"
	"unsafe"
	"time"
)

// Commit
type Commit struct {
	ptr *C.git_commit
}

// Id() and Type() satisfy Object
func (c *Commit) Id() *Oid {
	return newOidFromC(C.git_commit_id(c.ptr))
}

func (c *Commit) Type() int {
	return OBJ_COMMIT
}

func (c *Commit) Message() string {
	return C.GoString(C.git_commit_message(c.ptr))
}

func (c *Commit) Free() {
	runtime.SetFinalizer(c, nil)
	C.git_commit_free(c.ptr)
}

func newCommitFromC(ptr *C.git_commit) *Commit {
	commit := &Commit{ptr}
	runtime.SetFinalizer(commit, (*Commit).Free)

	return commit
}

func (c *Commit) Tree() (*Tree, error) {
	var ptr *C.git_tree

	err := C.git_commit_tree(&ptr, c.ptr)
	if err < 0 {
		return nil, LastError()
	}

	return newTreeFromC(ptr), nil
}

func (c *Commit) TreeId() *Oid {
	return newOidFromC(C.git_commit_tree_id(c.ptr))
}

func (c *Commit) Author() *Signature {
	ptr := C.git_commit_author(c.ptr)
	return newSignatureFromC(ptr)
}

func (c *Commit) Committer() *Signature {
	ptr := C.git_commit_committer(c.ptr)
	return newSignatureFromC(ptr)
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
	return offset/60
}

func (sig *Signature) toC() (*C.git_signature) {
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
