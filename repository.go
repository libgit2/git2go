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

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_repository_open(&repo.ptr, cpath)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(repo, (*Repository).Free)
	return repo, nil
}

func InitRepository(path string, isbare bool) (*Repository, error) {
	repo := new(Repository)

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_repository_init(&repo.ptr, cpath, ucbool(isbare))
	if ret < 0 {
		return nil, MakeGitError(ret)
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

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_repository_config(&config.ptr, v.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(config, (*Config).Free)
	return config, nil
}

func (v *Repository) Index() (*Index, error) {
	var ptr *C.git_index

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_repository_index(&ptr, v.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newIndexFromC(ptr), nil
}

func (v *Repository) lookupType(id *Oid, t ObjectType) (Object, error) {
	var ptr *C.git_object

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_object_lookup(&ptr, v.ptr, id.toC(), C.git_otype(t))
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return allocObject(ptr), nil
}

func (v *Repository) Lookup(id *Oid) (Object, error) {
	return v.lookupType(id, ObjectAny)
}

func (v *Repository) LookupTree(id *Oid) (*Tree, error) {
	obj, err := v.lookupType(id, ObjectTree)
	if err != nil {
		return nil, err
	}

	return obj.(*Tree), nil
}

func (v *Repository) LookupCommit(id *Oid) (*Commit, error) {
	obj, err := v.lookupType(id, ObjectCommit)
	if err != nil {
		return nil, err
	}

	return obj.(*Commit), nil
}

func (v *Repository) LookupBlob(id *Oid) (*Blob, error) {
	obj, err := v.lookupType(id, ObjectBlob)
	if err != nil {
		return nil, err
	}

	return obj.(*Blob), nil
}

func (v *Repository) LookupReference(name string) (*Reference, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var ptr *C.git_reference

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_reference_lookup(&ptr, v.ptr, cname)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newReferenceFromC(ptr), nil
}

func (v *Repository) CreateReference(name string, id *Oid, force bool, sig *Signature, msg string) (*Reference, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	csig := sig.toC()
	defer C.free(unsafe.Pointer(csig))

	var cmsg *C.char
	if msg == "" {
		cmsg = nil
	} else {
		cmsg = C.CString(msg)
		defer C.free(unsafe.Pointer(cmsg))
	}

	var ptr *C.git_reference

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_reference_create(&ptr, v.ptr, cname, id.toC(), cbool(force), csig, cmsg)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newReferenceFromC(ptr), nil
}

func (v *Repository) CreateSymbolicReference(name, target string, force bool, sig *Signature, msg string) (*Reference, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ctarget := C.CString(target)
	defer C.free(unsafe.Pointer(ctarget))

	csig := sig.toC()
	defer C.free(unsafe.Pointer(csig))

	var cmsg *C.char
	if msg == "" {
		cmsg = nil
	} else {
		cmsg = C.CString(msg)
		defer C.free(unsafe.Pointer(cmsg))
	}

	var ptr *C.git_reference

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_reference_symbolic_create(&ptr, v.ptr, cname, ctarget, cbool(force), csig, cmsg)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newReferenceFromC(ptr), nil
}

func (v *Repository) Walk() (*RevWalk, error) {
	walk := new(RevWalk)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_revwalk_new(&walk.ptr, v.ptr)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
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

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_commit_create(
		oid.toC(), v.ptr, cref,
		authorSig, committerSig,
		nil, cmsg, tree.ptr, C.size_t(nparents), parentsarg)

	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return oid, nil
}

func (v *Odb) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_odb_free(v.ptr)
}

func (v *Repository) Odb() (odb *Odb, err error) {
	odb = new(Odb)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if ret := C.git_repository_odb(&odb.ptr, v.ptr); ret < 0 {
		return nil, MakeGitError(ret)
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

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if ret := C.git_repository_set_workdir(repo.ptr, cstr, cbool(updateGitlink)); ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

func (v *Repository) TreeBuilder() (*TreeBuilder, error) {
	bld := new(TreeBuilder)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if ret := C.git_treebuilder_create(&bld.ptr, nil); ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(bld, (*TreeBuilder).Free)

	bld.repo = v
	return bld, nil
}

func (v *Repository) RevparseSingle(spec string) (Object, error) {
	cspec := C.CString(spec)
	defer C.free(unsafe.Pointer(cspec))

	var ptr *C.git_object

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_revparse_single(&ptr, v.ptr, cspec)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return allocObject(ptr), nil
}

// EnsureLog ensures that there is a reflog for the given reference
// name and creates an empty one if necessary.
func (v *Repository) EnsureLog(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_ensure_log(v.ptr, cname)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

// HasLog returns whether there is a reflog for the given reference
// name
func (v *Repository) HasLog(name string) (bool, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_has_log(v.ptr, cname)
	if ret < 0 {
		return false, MakeGitError(ret)
	}

	return ret == 1, nil
}

// DwimReference looks up a reference by DWIMing its short name
func (v *Repository) DwimReference(name string) (*Reference, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_reference
	ret := C.git_reference_dwim(&ptr, v.ptr, cname)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReferenceFromC(ptr), nil
}
