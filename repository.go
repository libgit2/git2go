package git

/*
#include <git2.h>
#include <git2/sys/repository.h>
#include <git2/sys/commit.h>
#include <string.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

// Repository
type Repository struct {
	ptr *C.git_repository
	// Remotes represents the collection of remotes and can be
	// used to add, remove and configure remotes for this
	// repository.
	Remotes RemoteCollection
	// Submodules represents the collection of submodules and can
	// be used to add, remove and configure submodules in this
	// repository.
	Submodules SubmoduleCollection
	// References represents the collection of references and can
	// be used to create, remove or update references for this repository.
	References ReferenceCollection
	// Notes represents the collection of notes and can be used to
	// read, write and delete notes from this repository.
	Notes NoteCollection
	// Tags represents the collection of tags and can be used to create,
	// list, iterate and remove tags in this repository.
	Tags TagsCollection
	// Stashes represents the collection of stashes and can be used to
	// save, apply and iterate over stash states in this repository.
	Stashes StashCollection
}

func newRepositoryFromC(ptr *C.git_repository) *Repository {
	repo := &Repository{ptr: ptr}

	repo.Remotes.repo = repo
	repo.Submodules.repo = repo
	repo.References.repo = repo
	repo.Notes.repo = repo
	repo.Tags.repo = repo
	repo.Stashes.repo = repo

	runtime.SetFinalizer(repo, (*Repository).Free)

	return repo
}

func OpenRepository(path string) (*Repository, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_repository
	ret := C.git_repository_open(&ptr, cpath)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newRepositoryFromC(ptr), nil
}

type RepositoryOpenFlag int

const (
	RepositoryOpenNoSearch RepositoryOpenFlag = C.GIT_REPOSITORY_OPEN_NO_SEARCH
	RepositoryOpenCrossFs  RepositoryOpenFlag = C.GIT_REPOSITORY_OPEN_CROSS_FS
	RepositoryOpenBare     RepositoryOpenFlag = C.GIT_REPOSITORY_OPEN_BARE
	RepositoryOpenFromEnv  RepositoryOpenFlag = C.GIT_REPOSITORY_OPEN_FROM_ENV
	RepositoryOpenNoDotGit RepositoryOpenFlag = C.GIT_REPOSITORY_OPEN_NO_DOTGIT
)

func OpenRepositoryExtended(path string, flags RepositoryOpenFlag, ceiling string) (*Repository, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	var cceiling *C.char = nil
	if len(ceiling) > 0 {
		cceiling = C.CString(ceiling)
		defer C.free(unsafe.Pointer(cceiling))
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_repository
	ret := C.git_repository_open_ext(&ptr, cpath, C.uint(flags), cceiling)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newRepositoryFromC(ptr), nil
}

func InitRepository(path string, isbare bool) (*Repository, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_repository
	ret := C.git_repository_init(&ptr, cpath, ucbool(isbare))
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newRepositoryFromC(ptr), nil
}

func NewRepositoryWrapOdb(odb *Odb) (repo *Repository, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_repository
	ret := C.git_repository_wrap_odb(&ptr, odb.ptr)
	runtime.KeepAlive(odb)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newRepositoryFromC(ptr), nil
}

func (v *Repository) SetRefdb(refdb *Refdb) {
	C.git_repository_set_refdb(v.ptr, refdb.ptr)
	runtime.KeepAlive(v)
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
	runtime.KeepAlive(v)
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

	return newIndexFromC(ptr, v), nil
}

func (v *Repository) lookupType(id *Oid, t ObjectType) (*Object, error) {
	var ptr *C.git_object

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_object_lookup(&ptr, v.ptr, id.toC(), C.git_object_t(t))
	runtime.KeepAlive(id)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return allocObject(ptr, v), nil
}

func (v *Repository) lookupPrefixType(id *Oid, prefix uint, t ObjectType) (*Object, error) {
	var ptr *C.git_object

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_object_lookup_prefix(&ptr, v.ptr, id.toC(), C.size_t(prefix), C.git_object_t(t))
	runtime.KeepAlive(id)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return allocObject(ptr, v), nil
}

func (v *Repository) Lookup(id *Oid) (*Object, error) {
	return v.lookupType(id, ObjectAny)
}

// LookupPrefix looks up an object by its OID given a prefix of its identifier.
func (v *Repository) LookupPrefix(id *Oid, prefix uint) (*Object, error) {
	return v.lookupPrefixType(id, prefix, ObjectAny)
}

func (v *Repository) LookupTree(id *Oid) (*Tree, error) {
	obj, err := v.lookupType(id, ObjectTree)
	if err != nil {
		return nil, err
	}
	defer obj.Free()

	return obj.AsTree()
}

// LookupPrefixTree looks up a tree by its OID given a prefix of its identifier.
func (v *Repository) LookupPrefixTree(id *Oid, prefix uint) (*Tree, error) {
	obj, err := v.lookupPrefixType(id, prefix, ObjectTree)
	if err != nil {
		return nil, err
	}
	defer obj.Free()

	return obj.AsTree()
}

func (v *Repository) LookupCommit(id *Oid) (*Commit, error) {
	obj, err := v.lookupType(id, ObjectCommit)
	if err != nil {
		return nil, err
	}
	defer obj.Free()

	return obj.AsCommit()
}

// LookupPrefixCommit looks up a commit by its OID given a prefix of its identifier.
func (v *Repository) LookupPrefixCommit(id *Oid, prefix uint) (*Commit, error) {
	obj, err := v.lookupPrefixType(id, prefix, ObjectCommit)
	if err != nil {
		return nil, err
	}
	defer obj.Free()

	return obj.AsCommit()
}

func (v *Repository) LookupBlob(id *Oid) (*Blob, error) {
	obj, err := v.lookupType(id, ObjectBlob)
	if err != nil {
		return nil, err
	}
	defer obj.Free()

	return obj.AsBlob()
}

// LookupPrefixBlob looks up a blob by its OID given a prefix of its identifier.
func (v *Repository) LookupPrefixBlob(id *Oid, prefix uint) (*Blob, error) {
	obj, err := v.lookupPrefixType(id, prefix, ObjectBlob)
	if err != nil {
		return nil, err
	}
	defer obj.Free()

	return obj.AsBlob()
}

func (v *Repository) LookupTag(id *Oid) (*Tag, error) {
	obj, err := v.lookupType(id, ObjectTag)
	if err != nil {
		return nil, err
	}
	defer obj.Free()

	return obj.AsTag()
}

// LookupPrefixTag looks up a tag by its OID given a prefix of its identifier.
func (v *Repository) LookupPrefixTag(id *Oid, prefix uint) (*Tag, error) {
	obj, err := v.lookupPrefixType(id, prefix, ObjectTag)
	if err != nil {
		return nil, err
	}
	defer obj.Free()

	return obj.AsTag()
}

func (v *Repository) Head() (*Reference, error) {
	var ptr *C.git_reference

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_repository_head(&ptr, v.ptr)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newReferenceFromC(ptr, v), nil
}

func (v *Repository) SetHead(refname string) error {
	cname := C.CString(refname)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_repository_set_head(v.ptr, cname)
	runtime.KeepAlive(v)
	if ecode != 0 {
		return MakeGitError(ecode)
	}
	return nil
}

func (v *Repository) SetHeadDetached(id *Oid) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_repository_set_head_detached(v.ptr, id.toC())
	runtime.KeepAlive(v)
	runtime.KeepAlive(id)
	if ecode != 0 {
		return MakeGitError(ecode)
	}
	return nil
}

func (v *Repository) IsHeadDetached() (bool, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_repository_head_detached(v.ptr)
	runtime.KeepAlive(v)
	if ret < 0 {
		return false, MakeGitError(ret)
	}

	return ret != 0, nil
}

func (v *Repository) IsHeadUnborn() (bool, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_repository_head_unborn(v.ptr)
	runtime.KeepAlive(v)
	if ret < 0 {
		return false, MakeGitError(ret)
	}
	return ret != 0, nil
}

func (v *Repository) IsEmpty() (bool, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_repository_is_empty(v.ptr)
	runtime.KeepAlive(v)
	if ret < 0 {
		return false, MakeGitError(ret)
	}

	return ret != 0, nil
}

func (v *Repository) IsShallow() (bool, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_repository_is_shallow(v.ptr)
	runtime.KeepAlive(v)
	if ret < 0 {
		return false, MakeGitError(ret)
	}
	return ret != 0, nil
}

func (v *Repository) Walk() (*RevWalk, error) {

	var walkPtr *C.git_revwalk

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_revwalk_new(&walkPtr, v.ptr)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return revWalkFromC(v, walkPtr), nil
}

func (v *Repository) CreateCommit(
	refname string, author, committer *Signature,
	message string, tree *Tree, parents ...*Commit) (*Oid, error) {

	oid := new(Oid)

	var cref *C.char
	if refname == "" {
		cref = nil
	} else {
		cref = C.CString(refname)
		defer C.free(unsafe.Pointer(cref))
	}

	cmsg := C.CString(message)
	defer C.free(unsafe.Pointer(cmsg))

	var cparents []*C.git_commit = nil
	var parentsarg **C.git_commit = nil

	nparents := len(parents)
	if nparents > 0 {
		cparents = make([]*C.git_commit, nparents)
		for i, v := range parents {
			cparents[i] = v.cast_ptr
		}
		parentsarg = &cparents[0]
	}

	authorSig, err := author.toC()
	if err != nil {
		return nil, err
	}
	defer C.git_signature_free(authorSig)

	committerSig, err := committer.toC()
	if err != nil {
		return nil, err
	}
	defer C.git_signature_free(committerSig)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_commit_create(
		oid.toC(), v.ptr, cref,
		authorSig, committerSig,
		nil, cmsg, tree.cast_ptr, C.size_t(nparents), parentsarg)

	runtime.KeepAlive(v)
	runtime.KeepAlive(oid)
	runtime.KeepAlive(parents)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return oid, nil
}

func (v *Repository) CreateCommitFromIds(
	refname string, author, committer *Signature,
	message string, tree *Oid, parents ...*Oid) (*Oid, error) {

	oid := new(Oid)

	var cref *C.char
	if refname == "" {
		cref = nil
	} else {
		cref = C.CString(refname)
		defer C.free(unsafe.Pointer(cref))
	}

	cmsg := C.CString(message)
	defer C.free(unsafe.Pointer(cmsg))

	var parentsarg **C.git_oid = nil

	nparents := len(parents)
	if nparents > 0 {
		// All this awful pointer arithmetic is needed to avoid passing a Go
		// pointer to Go pointer into C. Other methods (like CreateCommits) are
		// fine without this workaround because they are just passing Go pointers
		// to C pointers, but arrays-of-pointers-to-git_oid are a bit special since
		// both the array and the objects are allocated from Go.
		var emptyOidPtr *C.git_oid
		sizeofOidPtr := unsafe.Sizeof(emptyOidPtr)
		parentsarg = (**C.git_oid)(C.calloc(C.size_t(uintptr(nparents)), C.size_t(sizeofOidPtr)))
		defer C.free(unsafe.Pointer(parentsarg))
		parentsptr := uintptr(unsafe.Pointer(parentsarg))
		for _, v := range parents {
			*(**C.git_oid)(unsafe.Pointer(parentsptr)) = v.toC()
			parentsptr += sizeofOidPtr
		}
	}

	authorSig, err := author.toC()
	if err != nil {
		return nil, err
	}
	defer C.git_signature_free(authorSig)

	committerSig, err := committer.toC()
	if err != nil {
		return nil, err
	}
	defer C.git_signature_free(committerSig)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_commit_create_from_ids(
		oid.toC(), v.ptr, cref,
		authorSig, committerSig,
		nil, cmsg, tree.toC(), C.size_t(nparents), parentsarg)

	runtime.KeepAlive(v)
	runtime.KeepAlive(oid)
	runtime.KeepAlive(tree)
	runtime.KeepAlive(parents)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return oid, nil
}

func (v *Odb) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_odb_free(v.ptr)
}

func (v *Refdb) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_refdb_free(v.ptr)
}

func (v *Repository) Odb() (odb *Odb, err error) {
	odb = new(Odb)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_repository_odb(&odb.ptr, v.ptr)
	runtime.KeepAlive(v)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(odb, (*Odb).Free)
	return odb, nil
}

func (repo *Repository) Path() string {
	s := C.GoString(C.git_repository_path(repo.ptr))
	runtime.KeepAlive(repo)
	return s
}

func (repo *Repository) IsBare() bool {
	ret := C.git_repository_is_bare(repo.ptr) != 0
	runtime.KeepAlive(repo)
	return ret
}

func (repo *Repository) Workdir() string {
	s := C.GoString(C.git_repository_workdir(repo.ptr))
	runtime.KeepAlive(repo)
	return s
}

func (repo *Repository) SetWorkdir(workdir string, updateGitlink bool) error {
	cstr := C.CString(workdir)
	defer C.free(unsafe.Pointer(cstr))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if ret := C.git_repository_set_workdir(repo.ptr, cstr, cbool(updateGitlink)); ret < 0 {
		return MakeGitError(ret)
	}
	runtime.KeepAlive(repo)

	return nil
}

func (v *Repository) TreeBuilder() (*TreeBuilder, error) {
	bld := new(TreeBuilder)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if ret := C.git_treebuilder_new(&bld.ptr, v.ptr, nil); ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(bld, (*TreeBuilder).Free)

	bld.repo = v
	return bld, nil
}

func (v *Repository) TreeBuilderFromTree(tree *Tree) (*TreeBuilder, error) {
	bld := new(TreeBuilder)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if ret := C.git_treebuilder_new(&bld.ptr, v.ptr, tree.cast_ptr); ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(bld, (*TreeBuilder).Free)

	bld.repo = v
	return bld, nil
}

type RepositoryState int

const (
	RepositoryStateNone                 RepositoryState = C.GIT_REPOSITORY_STATE_NONE
	RepositoryStateMerge                RepositoryState = C.GIT_REPOSITORY_STATE_MERGE
	RepositoryStateRevert               RepositoryState = C.GIT_REPOSITORY_STATE_REVERT
	RepositoryStateCherrypick           RepositoryState = C.GIT_REPOSITORY_STATE_CHERRYPICK
	RepositoryStateBisect               RepositoryState = C.GIT_REPOSITORY_STATE_BISECT
	RepositoryStateRebase               RepositoryState = C.GIT_REPOSITORY_STATE_REBASE
	RepositoryStateRebaseInteractive    RepositoryState = C.GIT_REPOSITORY_STATE_REBASE_INTERACTIVE
	RepositoryStateRebaseMerge          RepositoryState = C.GIT_REPOSITORY_STATE_REBASE_MERGE
	RepositoryStateApplyMailbox         RepositoryState = C.GIT_REPOSITORY_STATE_APPLY_MAILBOX
	RepositoryStateApplyMailboxOrRebase RepositoryState = C.GIT_REPOSITORY_STATE_APPLY_MAILBOX_OR_REBASE
)

func (r *Repository) State() RepositoryState {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := RepositoryState(C.git_repository_state(r.ptr))
	runtime.KeepAlive(r)

	return ret
}

func (r *Repository) StateCleanup() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cErr := C.git_repository_state_cleanup(r.ptr)
	runtime.KeepAlive(r)
	if cErr < 0 {
		return MakeGitError(cErr)
	}
	return nil
}

func (r *Repository) AddGitIgnoreRules(rules string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	crules := C.CString(rules)
	defer C.free(unsafe.Pointer(crules))
	ret := C.git_ignore_add_rule(r.ptr, crules)
	runtime.KeepAlive(r)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (r *Repository) ClearGitIgnoreRules() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_ignore_clear_internal_rules(r.ptr)
	runtime.KeepAlive(r)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}
