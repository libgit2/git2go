package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
*/
import "C"
import "runtime"

type ObjectType int

var (
	OBJ_ANY ObjectType = C.GIT_OBJ_ANY
	OBJ_BAD ObjectType = C.GIT_OBJ_BAD
	OBJ_COMMIT ObjectType = C.GIT_OBJ_COMMIT
	OBJ_TREE ObjectType = C.GIT_OBJ_TREE
	OBJ_BLOB ObjectType = C.GIT_OBJ_BLOB
	OBJ_TAG ObjectType = C.GIT_OBJ_TAG
)

type Object interface {
	Free()
	Id() *Oid
	Type() ObjectType
}

type gitObject struct {
	ptr *C.git_object
}

func (t ObjectType) String() (string) {
	switch (t) {
	case OBJ_ANY:
		return "Any"
	case OBJ_BAD:
		return "Bad"
	case OBJ_COMMIT:
		return "Commit"
	case OBJ_TREE:
		return "Tree"
	case OBJ_BLOB:
		return "Blob"
	case OBJ_TAG:
		return "tag"
	}
	// Never reached
	return ""
}

func (o gitObject) Id() *Oid {
	return newOidFromC(C.git_commit_id(o.ptr))
}

func (o gitObject) Type() ObjectType {
	return ObjectType(C.git_object_type(o.ptr))
}

func (o gitObject) Free() {
	runtime.SetFinalizer(o, nil)
	C.git_commit_free(o.ptr)
}

func allocObject(cobj *C.git_object) Object {

	switch ObjectType(C.git_object_type(cobj)) {
	case OBJ_COMMIT:
		commit := &Commit{gitObject{cobj}}
		runtime.SetFinalizer(commit, (*Commit).Free)
		return commit

	case OBJ_TREE:
		tree := &Tree{gitObject{cobj}}
		runtime.SetFinalizer(tree, (*Tree).Free)
		return tree

	case OBJ_BLOB:
		blob := &Blob{gitObject{cobj}}
		runtime.SetFinalizer(blob, (*Blob).Free)
		return blob
	}

	return nil
}
