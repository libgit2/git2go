package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
*/
import "C"
import (
	"unsafe"
	"runtime"
)

// Repository
type Repository struct {
	ptr *C.git_repository
}

func Open(path string) (*Repository, error) {
	repo := new(Repository)

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	ret := C.git_repository_open(&repo.ptr, cpath)
	if ret < 0 {
		return nil, LastError()
	}

	runtime.SetFinalizer(repo, freeRepository)
	return repo, nil
}

func Init(path string, isbare bool) (*Repository, error) {
	repo := new(Repository)

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	ret := C.git_repository_init(&repo.ptr, cpath, ucbool(isbare))
	if ret < 0 {
		return nil, LastError()
	}

	runtime.SetFinalizer(repo, freeRepository)
	return repo, nil
}

func freeRepository(repo *Repository) {
	C.git_repository_free(repo.ptr)
}

func (v *Repository) Config() (*Config, error) {
	config := new(Config)

	ret := C.git_repository_config(&config.ptr, v.ptr)
	if ret < 0 {
		return nil, LastError()
	}

	return config, nil
}

func (v *Repository) LookupTree(oid *Oid) (*Tree, error) {
	tree := new(Tree)
	ret := C.git_tree_lookup(&tree.ptr, v.ptr, oid.toC())
	if ret < 0 {
		return nil, LastError()
	}

	return tree, nil
}

func (v *Repository) LookupCommit(o *Oid) (*Commit, error) {
	commit := new(Commit)
	ecode := C.git_commit_lookup(&commit.ptr, v.ptr, o.toC())
	if ecode < 0 {
		return nil, LastError()
	}

	return commit, nil
}

func (v *Repository) Walk() (*RevWalk, error) {
	walk := new(RevWalk)
	ecode := C.git_revwalk_new(&walk.ptr, v.ptr)
	if ecode < 0 {
		return nil, LastError()
	}

	walk.repo = v
	runtime.SetFinalizer(walk, freeRevWalk)
	return walk, nil
}

/* TODO 
func (v *Repository) Commit(
	refname string, author, committer *Signature,
	message string, tree *Tree, parents ...*Commit) (*Oid, error) {

	oid := new(Oid)
	cref := C.CString(refname)
	defer C.free(unsafe.Pointer(cref))
	cmsg := C.CString(message)
	defer C.free(unsafe.Pointer(cmsg))
	nparents := len(parents)
	var cparents []*C.git_commit = nil
	var parentsarg **C.git_commit = nil
	if nparents > 0 {
		cparents = make([]*C.git_commit, nparents)
		for i, v := range parents {
			cparents[i] = v.ptr
		}
		parentsarg = &cparents[0]
	}

	ret := C.git_commit_create(
		oid.toC(), v.ptr, cref,
		author.git_signature, committer.git_signature,
		nil, cmsg, tree.ptr, C.int(nparents), parentsarg)

	if ret < GIT_SUCCESS {
		return nil, LastError()
	}

	return oid, nil
}
*/

func freeOdb(odb *Odb) {
	C.git_odb_free(odb.ptr)
}

func (v *Repository) Odb() (odb *Odb, err error) {
	odb = new(Odb)
	if ret := C.git_repository_odb(&odb.ptr, v.ptr); ret < 0 {
		return nil, LastError()
	}

	runtime.SetFinalizer(odb, freeOdb)
	return
}

func (v *Repository) TreeBuilder() (*TreeBuilder, error) {
	bld := new(TreeBuilder)
	if ret := C.git_treebuilder_create(&bld.ptr, nil); ret < 0 {
		return nil, LastError()
	}

	bld.repo = v
	return bld, nil
}

