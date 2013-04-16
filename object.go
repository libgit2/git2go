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

func allocObject(cobj *C.git_object) Object {
	var object Object

	switch ObjectType(C.git_object_type(cobj)) {
	case OBJ_COMMIT:
		object = &Commit{cobj}
		runtime.SetFinalizer(object, (*Commit).Free)

	case OBJ_TREE:
		object = &Tree{cobj}
		runtime.SetFinalizer(object, (*Tree).Free)

	case OBJ_BLOB:
		object = &Blob{cobj}
		runtime.SetFinalizer(object, (*Blob).Free)

	default:
		return nil
	}

	return object
}
