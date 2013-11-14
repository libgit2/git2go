package git

/*
#include <git2.h>
#include <git2/errors.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

// Repository
type Repository struct {
	ptr *C.git_repository
}

func OpenRepository(path string) (*Repository, error) {
	repo := new(Repository)

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	ret := C.git_repository_open(&repo.ptr, cpath)
	if ret < 0 {
		return nil, LastError()
	}

	runtime.SetFinalizer(repo, (*Repository).Free)
	return repo, nil
}

func InitRepository(path string, isbare bool) (*Repository, error) {
	repo := new(Repository)

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	ret := C.git_repository_init(&repo.ptr, cpath, ucbool(isbare))
	if ret < 0 {
		return nil, LastError()
	}

	runtime.SetFinalizer(repo, (*Repository).Free)
	return repo, nil
}

func (v *Repository) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_repository_free(v.ptr)
}

func (v *Repository) Config() (*Config, error) {
	config := new(Config)

	ret := C.git_repository_config(&config.ptr, v.ptr)
	if ret < 0 {
		return nil, LastError()
	}

	return config, nil
}

func (v *Repository) Index() (*Index, error) {
	var ptr *C.git_index
	ret := C.git_repository_index(&ptr, v.ptr)
	if ret < 0 {
		return nil, LastError()
	}

	return newIndexFromC(ptr), nil
}

func (v *Repository) lookupType(oid *Oid, t ObjectType) (Object, error) {
	var ptr *C.git_object
	ret := C.git_object_lookup(&ptr, v.ptr, oid.toC(), C.git_otype(t))
	if ret < 0 {
		return nil, LastError()
	}

	return allocObject(ptr), nil
}

func (v *Repository) Lookup(oid *Oid) (Object, error) {
	return v.lookupType(oid, ObjectAny)
}

func (v *Repository) LookupTree(oid *Oid) (*Tree, error) {
	obj, err := v.lookupType(oid, ObjectTree)
	if err != nil {
		return nil, err
	}

	return obj.(*Tree), nil
}

func (v *Repository) LookupCommit(oid *Oid) (*Commit, error) {
	obj, err := v.lookupType(oid, ObjectCommit)
	if err != nil {
		return nil, err
	}

	return obj.(*Commit), nil
}

func (v *Repository) LookupBlob(oid *Oid) (*Blob, error) {
	obj, err := v.lookupType(oid, ObjectBlob)
	if err != nil {
		return nil, err
	}

	return obj.(*Blob), nil
}

func (v *Repository) LookupReference(name string) (*Reference, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var ptr *C.git_reference

	ecode := C.git_reference_lookup(&ptr, v.ptr, cname)
	if ecode < 0 {
		return nil, LastError()
	}

	return newReferenceFromC(ptr), nil
}

func (v *Repository) CreateReference(name string, oid *Oid, force bool) (*Reference, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var ptr *C.git_reference

	ecode := C.git_reference_create(&ptr, v.ptr, cname, oid.toC(), cbool(force))
	if ecode < 0 {
		return nil, LastError()
	}

	return newReferenceFromC(ptr), nil
}

func (v *Repository) CreateSymbolicReference(name, target string, force bool) (*Reference, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ctarget := C.CString(target)
	defer C.free(unsafe.Pointer(ctarget))
	var ptr *C.git_reference

	ecode := C.git_reference_symbolic_create(&ptr, v.ptr, cname, ctarget, cbool(force))
	if ecode < 0 {
		return nil, LastError()
	}

	return newReferenceFromC(ptr), nil
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

func (v *Repository) CreateCommit(
	refname string, author, committer *Signature,
	message string, tree *Tree, parents ...*Commit) (*Oid, error) {

	oid := new(Oid)

	cref := C.CString(refname)
	defer C.free(unsafe.Pointer(cref))

	cmsg := C.CString(message)
	defer C.free(unsafe.Pointer(cmsg))

	var cparents []*C.git_commit = nil
	var parentsarg **C.git_commit = nil

	nparents := len(parents)
	if nparents > 0 {
		cparents = make([]*C.git_commit, nparents)
		for i, v := range parents {
			cparents[i] = v.ptr
		}
		parentsarg = &cparents[0]
	}

	authorSig := author.toC()
	defer C.git_signature_free(authorSig)

	committerSig := committer.toC()
	defer C.git_signature_free(committerSig)

	ret := C.git_commit_create(
		oid.toC(), v.ptr, cref,
		authorSig, committerSig,
		nil, cmsg, tree.ptr, C.int(nparents), parentsarg)

	if ret < 0 {
		return nil, LastError()
	}

	return oid, nil
}

func (v *Odb) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_odb_free(v.ptr)
}

func (v *Repository) Odb() (odb *Odb, err error) {
	odb = new(Odb)
	if ret := C.git_repository_odb(&odb.ptr, v.ptr); ret < 0 {
		return nil, LastError()
	}

	runtime.SetFinalizer(odb, (*Odb).Free)
	return
}

func (repo *Repository) Path() string {
	return C.GoString(C.git_repository_path(repo.ptr))
}

func (repo *Repository) IsBare() bool {
	return C.git_repository_is_bare(repo.ptr) != 0
}

func (repo *Repository) Workdir() string {
	return C.GoString(C.git_repository_workdir(repo.ptr))
}

func (repo *Repository) SetWorkdir(workdir string, updateGitlink bool) error {
	cstr := C.CString(workdir)
	defer C.free(unsafe.Pointer(cstr))

	if C.git_repository_set_workdir(repo.ptr, cstr, cbool(updateGitlink)) < 0 {
		return LastError()
	}
	return nil
}

func (v *Repository) TreeBuilder() (*TreeBuilder, error) {
	bld := new(TreeBuilder)
	if ret := C.git_treebuilder_create(&bld.ptr, nil); ret < 0 {
		return nil, LastError()
	}
	runtime.SetFinalizer(bld, (*TreeBuilder).Free)

	bld.repo = v
	return bld, nil
}

func (v *Repository) RevparseSingle(spec string) (Object, error) {
	cspec := C.CString(spec)
	defer C.free(unsafe.Pointer(cspec))

	var ptr *C.git_object
	ecode := C.git_revparse_single(&ptr, v.ptr, cspec)
	if ecode < 0 {
		return nil, LastError()
	}

	return allocObject(ptr), nil
}
