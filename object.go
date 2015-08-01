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
	Peel(t ObjectType) (Object, error)
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

// Peel recursively peels an object until an object of the specified type is met.
//
// If the query cannot be satisfied due to the object model, ErrInvalidSpec
// will be returned (e.g. trying to peel a blob to a tree).
//
// If you pass ObjectAny as the target type, then the object will be peeled
// until the type changes. A tag will be peeled until the referenced object
// is no longer a tag, and a commit will be peeled to a tree. Any other object
// type will return ErrInvalidSpec.
//
// If peeling a tag we discover an object which cannot be peeled to the target
// type due to the object model, an error will be returned.
func (o *gitObject) Peel(t ObjectType) (Object, error) {
	var cobj *C.git_object

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.git_object_peel(&cobj, o.ptr, C.git_otype(t)); err < 0 {
		return nil, MakeGitError(err)
	}

	return allocObject(cobj, o.repo), nil
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
