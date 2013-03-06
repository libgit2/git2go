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

func (c *Commit) Id() *Oid {
	return newOidFromC(C.git_commit_id(c.ptr))
}

func (c *Commit) Message() string {
	return C.GoString(C.git_commit_message(c.ptr))
}

func (c *Commit) Tree() (*Tree, error) {
	tree := new(Tree)

	err := C.git_commit_tree(&tree.ptr, c.ptr)
	if err < 0 {
		return nil, LastError()
	}

	runtime.SetFinalizer(tree, (*Tree).Free)
	return tree, nil
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
	Name string
	Email string
	UnixTime int64
	Offset int
}

func newSignatureFromC(sig *C.git_signature) *Signature {
	return &Signature{
		C.GoString(sig.name),
		C.GoString(sig.email),
		int64(sig.when.time),
		int(sig.when.offset),
	}
}

func (sig *Signature) Time() time.Time {
	loc := time.FixedZone("", sig.Offset*60)
	return time.Unix(sig.UnixTime, 0).In(loc)
}

func (sig *Signature) toC() (*C.git_signature) {
	var out *C.git_signature

	name := C.CString(sig.Name)
	defer C.free(unsafe.Pointer(name))

	email := C.CString(sig.Email)
	defer C.free(unsafe.Pointer(email))

	ret := C.git_signature_new(&out, name, email, C.git_time_t(sig.UnixTime), C.int(sig.Offset))
	if ret < 0 {
		return nil
	}

	return out
}
