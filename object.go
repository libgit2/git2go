package git

/*
#include <git2.h>
*/
import "C"
import "runtime"

type ObjectType int

const (
	ObjectAny    ObjectType = C.GIT_OBJ_ANY
	ObjectBad    ObjectType = C.GIT_OBJ_BAD
	ObjectCommit ObjectType = C.GIT_OBJ_COMMIT
	ObjectTree   ObjectType = C.GIT_OBJ_TREE
	ObjectBlob   ObjectType = C.GIT_OBJ_BLOB
	ObjectTag    ObjectType = C.GIT_OBJ_TAG
)

type Object interface {
	Free()
	Id() *Oid
	Type() ObjectType
	Owner() *Repository
}

type gitObject struct {
	ptr  *C.git_object
	repo *Repository
}

func (t ObjectType) String() string {
	switch t {
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

// Owner returns a weak reference to the repository which owns this
// object
func (o gitObject) Owner() *Repository {
	return &Repository{
		ptr: C.git_object_owner(o.ptr),
	}
}

func (o *gitObject) Free() {
	runtime.SetFinalizer(o, nil)
	C.git_object_free(o.ptr)
}

func allocObject(cobj *C.git_object, repo *Repository) Object {
	obj := gitObject{
		ptr:  cobj,
		repo: repo,
	}

	switch ObjectType(C.git_object_type(cobj)) {
	case ObjectCommit:
		commit := &Commit{
			gitObject: obj,
			cast_ptr:  (*C.git_commit)(cobj),
		}
		runtime.SetFinalizer(commit, (*Commit).Free)
		return commit

	case ObjectTree:
		tree := &Tree{
			gitObject: obj,
			cast_ptr:  (*C.git_tree)(cobj),
		}
		runtime.SetFinalizer(tree, (*Tree).Free)
		return tree

	case ObjectBlob:
		blob := &Blob{
			gitObject: obj,
			cast_ptr:  (*C.git_blob)(cobj),
		}
		runtime.SetFinalizer(blob, (*Blob).Free)
		return blob
	case ObjectTag:
		tag := &Tag{
			gitObject: obj,
			cast_ptr:  (*C.git_tag)(cobj),
		}
		runtime.SetFinalizer(tag, (*Tag).Free)
		return tag
	}

	return nil
}
