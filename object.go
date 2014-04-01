package git

/*
#include <git2.h>
#include <git2/errors.h>
*/
import "C"
import "runtime"

type ObjectType int

const (
	ObjectAny ObjectType = C.GIT_OBJ_ANY
	ObjectBad            = C.GIT_OBJ_BAD
	ObjectCommit         = C.GIT_OBJ_COMMIT
	ObjectTree           = C.GIT_OBJ_TREE
	ObjectBlob           = C.GIT_OBJ_BLOB
	ObjectTag            = C.GIT_OBJ_TAG
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
	case ObjectAny:
		return "Any"
	case ObjectBad:
		return "Bad"
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

func (o gitObject) Id() *Oid {
	return newOidFromC(C.git_object_id(o.ptr))
}

func (o gitObject) Type() ObjectType {
	return ObjectType(C.git_object_type(o.ptr))
}

func (o *gitObject) Free() {
	runtime.SetFinalizer(o, nil)
	C.git_object_free(o.ptr)
}

func allocObject(cobj *C.git_object) Object {

	switch ObjectType(C.git_object_type(cobj)) {
	case ObjectCommit:
		commit := &Commit{
			gitObject: gitObject{cobj},
			cast_ptr:  (*C.git_commit)(cobj),
		}
		runtime.SetFinalizer(commit, (*Commit).Free)
		return commit

	case ObjectTree:
		tree := &Tree{
			gitObject: gitObject{cobj},
			cast_ptr:  (*C.git_tree)(cobj),
		}
		runtime.SetFinalizer(tree, (*Tree).Free)
		return tree

	case ObjectBlob:
		blob := &Blob{
			gitObject: gitObject{cobj},
			cast_ptr:  (*C.git_blob)(cobj),
		}
		runtime.SetFinalizer(blob, (*Blob).Free)
		return blob
	}

	return nil
}
