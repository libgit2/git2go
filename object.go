package git

/*
#include <git2.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"runtime"
)

type ObjectType int

const (
	ObjectAny     ObjectType = C.GIT_OBJECT_ANY
	ObjectInvalid ObjectType = C.GIT_OBJECT_INVALID
	ObjectCommit  ObjectType = C.GIT_OBJECT_COMMIT
	ObjectTree    ObjectType = C.GIT_OBJECT_TREE
	ObjectBlob    ObjectType = C.GIT_OBJECT_BLOB
	ObjectTag     ObjectType = C.GIT_OBJECT_TAG
)

type Object struct {
	doNotCompare
	ptr  *C.git_object
	repo *Repository
}

// Objecter lets us accept any kind of Git object in functions.
type Objecter interface {
	AsObject() *Object
}

func (t ObjectType) String() string {
	switch t {
	case ObjectAny:
		return "Any"
	case ObjectInvalid:
		return "Invalid"
	case ObjectCommit:
		return "Commit"
	case ObjectTree:
		return "Tree"
	case ObjectBlob:
		return "Blob"
	case ObjectTag:
		return "Tag"
	}
	// Never reached
	return ""
}

func (o *Object) Id() *Oid {
	ret := newOidFromC(C.git_object_id(o.ptr))
	runtime.KeepAlive(o)
	return ret
}

func (o *Object) ShortId() (string, error) {
	resultBuf := C.git_buf{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_object_short_id(&resultBuf, o.ptr)
	runtime.KeepAlive(o)
	if ecode < 0 {
		return "", MakeGitError(ecode)
	}
	defer C.git_buf_dispose(&resultBuf)
	return C.GoString(resultBuf.ptr), nil
}

func (o *Object) Type() ObjectType {
	ret := ObjectType(C.git_object_type(o.ptr))
	runtime.KeepAlive(o)
	return ret
}

// Owner returns a weak reference to the repository which owns this object.
// This won't keep the underlying repository alive, but it should still be
// Freed.
func (o *Object) Owner() *Repository {
	repo := newRepositoryFromC(C.git_object_owner(o.ptr))
	runtime.KeepAlive(o)
	repo.weak = true
	return repo
}

func dupObject(obj *Object, kind ObjectType) (*C.git_object, error) {
	if obj.Type() != kind {
		return nil, errors.New(fmt.Sprintf("object is not a %v", kind))
	}

	var cobj *C.git_object

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C.git_object_dup(&cobj, obj.ptr)
	runtime.KeepAlive(obj)
	if err < 0 {
		return nil, MakeGitError(err)
	}

	return cobj, nil
}

func allocTree(ptr *C.git_tree, repo *Repository) *Tree {
	tree := &Tree{
		Object: Object{
			ptr:  (*C.git_object)(ptr),
			repo: repo,
		},
		cast_ptr: ptr,
	}
	runtime.SetFinalizer(tree, (*Tree).Free)

	return tree
}

func (o *Object) AsTree() (*Tree, error) {
	cobj, err := dupObject(o, ObjectTree)
	if err != nil {
		return nil, err
	}

	return allocTree((*C.git_tree)(cobj), o.repo), nil
}

func allocCommit(ptr *C.git_commit, repo *Repository) *Commit {
	commit := &Commit{
		Object: Object{
			ptr:  (*C.git_object)(ptr),
			repo: repo,
		},
		cast_ptr: ptr,
	}
	runtime.SetFinalizer(commit, (*Commit).Free)

	return commit
}

func (o *Object) AsCommit() (*Commit, error) {
	cobj, err := dupObject(o, ObjectCommit)
	if err != nil {
		return nil, err
	}

	return allocCommit((*C.git_commit)(cobj), o.repo), nil
}

func allocBlob(ptr *C.git_blob, repo *Repository) *Blob {
	blob := &Blob{
		Object: Object{
			ptr:  (*C.git_object)(ptr),
			repo: repo,
		},
		cast_ptr: ptr,
	}
	runtime.SetFinalizer(blob, (*Blob).Free)

	return blob
}

func (o *Object) AsBlob() (*Blob, error) {
	cobj, err := dupObject(o, ObjectBlob)
	if err != nil {
		return nil, err
	}

	return allocBlob((*C.git_blob)(cobj), o.repo), nil
}

func allocTag(ptr *C.git_tag, repo *Repository) *Tag {
	tag := &Tag{
		Object: Object{
			ptr:  (*C.git_object)(ptr),
			repo: repo,
		},
		cast_ptr: ptr,
	}
	runtime.SetFinalizer(tag, (*Tag).Free)

	return tag
}

func (o *Object) AsTag() (*Tag, error) {
	cobj, err := dupObject(o, ObjectTag)
	if err != nil {
		return nil, err
	}

	return allocTag((*C.git_tag)(cobj), o.repo), nil
}

func (o *Object) Free() {
	runtime.SetFinalizer(o, nil)
	C.git_object_free(o.ptr)
}

// Peel recursively peels an object until an object of the specified type is met.
//
// If the query cannot be satisfied due to the object model, ErrorCodeInvalidSpec
// will be returned (e.g. trying to peel a blob to a tree).
//
// If you pass ObjectAny as the target type, then the object will be peeled
// until the type changes. A tag will be peeled until the referenced object
// is no longer a tag, and a commit will be peeled to a tree. Any other object
// type will return ErrorCodeInvalidSpec.
//
// If peeling a tag we discover an object which cannot be peeled to the target
// type due to the object model, an error will be returned.
func (o *Object) Peel(t ObjectType) (*Object, error) {
	var cobj *C.git_object

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C.git_object_peel(&cobj, o.ptr, C.git_object_t(t))
	runtime.KeepAlive(o)
	if err < 0 {
		return nil, MakeGitError(err)
	}

	return allocObject(cobj, o.repo), nil
}

func allocObject(cobj *C.git_object, repo *Repository) *Object {
	obj := &Object{
		ptr:  cobj,
		repo: repo,
	}
	runtime.SetFinalizer(obj, (*Object).Free)

	return obj
}
