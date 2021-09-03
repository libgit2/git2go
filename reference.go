package git

/*
#include <git2.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type ReferenceType int

const (
	ReferenceSymbolic ReferenceType = C.GIT_REF_SYMBOLIC
	ReferenceOid      ReferenceType = C.GIT_REF_OID
)

type Reference struct {
	doNotCompare
	ptr  *C.git_reference
	repo *Repository
}

type ReferenceCollection struct {
	doNotCompare
	repo *Repository
}

func (c *ReferenceCollection) Lookup(name string) (*Reference, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var ptr *C.git_reference

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_reference_lookup(&ptr, c.repo.ptr, cname)
	runtime.KeepAlive(c)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newReferenceFromC(ptr, c.repo), nil
}

func (c *ReferenceCollection) Create(name string, id *Oid, force bool, msg string) (*Reference, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

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

	ecode := C.git_reference_create(&ptr, c.repo.ptr, cname, id.toC(), cbool(force), cmsg)
	runtime.KeepAlive(c)
	runtime.KeepAlive(id)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newReferenceFromC(ptr, c.repo), nil
}

func (c *ReferenceCollection) CreateSymbolic(name, target string, force bool, msg string) (*Reference, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ctarget := C.CString(target)
	defer C.free(unsafe.Pointer(ctarget))

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

	ecode := C.git_reference_symbolic_create(&ptr, c.repo.ptr, cname, ctarget, cbool(force), cmsg)
	runtime.KeepAlive(c)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newReferenceFromC(ptr, c.repo), nil
}

// EnsureLog ensures that there is a reflog for the given reference
// name and creates an empty one if necessary.
func (c *ReferenceCollection) EnsureLog(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_ensure_log(c.repo.ptr, cname)
	runtime.KeepAlive(c)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

// HasLog returns whether there is a reflog for the given reference
// name
func (c *ReferenceCollection) HasLog(name string) (bool, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_has_log(c.repo.ptr, cname)
	runtime.KeepAlive(c)
	if ret < 0 {
		return false, MakeGitError(ret)
	}

	return ret == 1, nil
}

// Dwim looks up a reference by DWIMing its short name
func (c *ReferenceCollection) Dwim(name string) (*Reference, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_reference
	ret := C.git_reference_dwim(&ptr, c.repo.ptr, cname)
	runtime.KeepAlive(c)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReferenceFromC(ptr, c.repo), nil
}

func newReferenceFromC(ptr *C.git_reference, repo *Repository) *Reference {
	ref := &Reference{ptr: ptr, repo: repo}
	runtime.SetFinalizer(ref, (*Reference).Free)
	return ref
}

func (v *Reference) SetSymbolicTarget(target string, msg string) (*Reference, error) {
	var ptr *C.git_reference

	ctarget := C.CString(target)
	defer C.free(unsafe.Pointer(ctarget))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var cmsg *C.char
	if msg == "" {
		cmsg = nil
	} else {
		cmsg = C.CString(msg)
		defer C.free(unsafe.Pointer(cmsg))
	}

	ret := C.git_reference_symbolic_set_target(&ptr, v.ptr, ctarget, cmsg)
	runtime.KeepAlive(v)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReferenceFromC(ptr, v.repo), nil
}

func (v *Reference) SetTarget(target *Oid, msg string) (*Reference, error) {
	var ptr *C.git_reference

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var cmsg *C.char
	if msg == "" {
		cmsg = nil
	} else {
		cmsg = C.CString(msg)
		defer C.free(unsafe.Pointer(cmsg))
	}

	ret := C.git_reference_set_target(&ptr, v.ptr, target.toC(), cmsg)
	runtime.KeepAlive(v)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReferenceFromC(ptr, v.repo), nil
}

func (v *Reference) Resolve() (*Reference, error) {
	var ptr *C.git_reference

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_resolve(&ptr, v.ptr)
	runtime.KeepAlive(v)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReferenceFromC(ptr, v.repo), nil
}

func (v *Reference) Rename(name string, force bool, msg string) (*Reference, error) {
	var ptr *C.git_reference
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var cmsg *C.char
	if msg == "" {
		cmsg = nil
	} else {
		cmsg = C.CString(msg)
		defer C.free(unsafe.Pointer(cmsg))
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_rename(&ptr, v.ptr, cname, cbool(force), cmsg)
	runtime.KeepAlive(v)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReferenceFromC(ptr, v.repo), nil
}

func (v *Reference) Target() *Oid {
	ret := newOidFromC(C.git_reference_target(v.ptr))
	runtime.KeepAlive(v)
	return ret
}

func (v *Reference) SymbolicTarget() string {
	var ret string
	cstr := C.git_reference_symbolic_target(v.ptr)

	if cstr != nil {
		return C.GoString(cstr)
	}

	runtime.KeepAlive(v)
	return ret
}

func (v *Reference) Delete() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_delete(v.ptr)
	runtime.KeepAlive(v)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

func (v *Reference) Peel(t ObjectType) (*Object, error) {
	var cobj *C.git_object

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C.git_reference_peel(&cobj, v.ptr, C.git_object_t(t))
	runtime.KeepAlive(v)
	if err < 0 {
		return nil, MakeGitError(err)
	}

	return allocObject(cobj, v.repo), nil
}

// Owner returns a weak reference to the repository which owns this reference.
// This won't keep the underlying repository alive, but it should still be
// Freed.
func (v *Reference) Owner() *Repository {
	repo := newRepositoryFromC(C.git_reference_owner(v.ptr))
	runtime.KeepAlive(v)
	repo.weak = true
	return repo
}

// Cmp compares v to ref2. It returns 0 on equality, otherwise a
// stable sorting.
func (v *Reference) Cmp(ref2 *Reference) int {
	ret := int(C.git_reference_cmp(v.ptr, ref2.ptr))
	runtime.KeepAlive(v)
	runtime.KeepAlive(ref2)
	return ret
}

// Shorthand returns a "human-readable" short reference name.
func (v *Reference) Shorthand() string {
	ret := C.GoString(C.git_reference_shorthand(v.ptr))
	runtime.KeepAlive(v)
	return ret
}

// Name returns the full name of v.
func (v *Reference) Name() string {
	ret := C.GoString(C.git_reference_name(v.ptr))
	runtime.KeepAlive(v)
	return ret
}

func (v *Reference) Type() ReferenceType {
	ret := ReferenceType(C.git_reference_type(v.ptr))
	runtime.KeepAlive(v)
	return ret
}

func (v *Reference) IsBranch() bool {
	ret := C.git_reference_is_branch(v.ptr) == 1
	runtime.KeepAlive(v)
	return ret
}

func (v *Reference) IsRemote() bool {
	ret := C.git_reference_is_remote(v.ptr) == 1
	runtime.KeepAlive(v)
	return ret
}

func (v *Reference) IsTag() bool {
	ret := C.git_reference_is_tag(v.ptr) == 1
	runtime.KeepAlive(v)
	return ret
}

// IsNote checks if the reference is a note.
func (v *Reference) IsNote() bool {
	ret := C.git_reference_is_note(v.ptr) == 1
	runtime.KeepAlive(v)
	return ret
}

func (v *Reference) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_reference_free(v.ptr)
}

type ReferenceIterator struct {
	doNotCompare
	ptr  *C.git_reference_iterator
	repo *Repository
}

type ReferenceNameIterator struct {
	doNotCompare
	*ReferenceIterator
}

// NewReferenceIterator creates a new iterator over reference names
func (repo *Repository) NewReferenceIterator() (*ReferenceIterator, error) {
	var ptr *C.git_reference_iterator

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_iterator_new(&ptr, repo.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReferenceIteratorFromC(ptr, repo), nil
}

// NewReferenceIterator creates a new branch iterator over reference names
func (repo *Repository) NewReferenceNameIterator() (*ReferenceNameIterator, error) {
	var ptr *C.git_reference_iterator

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_iterator_new(&ptr, repo.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	iter := newReferenceIteratorFromC(ptr, repo)
	return iter.Names(), nil
}

// NewReferenceIteratorGlob creates an iterator over reference names
// that match the speicified glob. The glob is of the usual fnmatch
// type.
func (repo *Repository) NewReferenceIteratorGlob(glob string) (*ReferenceIterator, error) {
	cstr := C.CString(glob)
	defer C.free(unsafe.Pointer(cstr))
	var ptr *C.git_reference_iterator

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_iterator_glob_new(&ptr, repo.ptr, cstr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReferenceIteratorFromC(ptr, repo), nil
}

func (i *ReferenceIterator) Names() *ReferenceNameIterator {
	return &ReferenceNameIterator{ReferenceIterator: i}
}

// NextName retrieves the next reference name. If the iteration is over,
// the returned error code is git.ErrorCodeIterOver
func (v *ReferenceNameIterator) Next() (string, error) {
	var ptr *C.char

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_next_name(&ptr, v.ptr)
	if ret < 0 {
		return "", MakeGitError(ret)
	}

	return C.GoString(ptr), nil
}

// Next retrieves the next reference. If the iterationis over, the
// returned error code is git.ErrorCodeIterOver
func (v *ReferenceIterator) Next() (*Reference, error) {
	var ptr *C.git_reference

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_next(&ptr, v.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReferenceFromC(ptr, v.repo), nil
}

func newReferenceIteratorFromC(ptr *C.git_reference_iterator, r *Repository) *ReferenceIterator {
	iter := &ReferenceIterator{
		ptr:  ptr,
		repo: r,
	}
	runtime.SetFinalizer(iter, (*ReferenceIterator).Free)
	return iter
}

// Free the reference iterator
func (v *ReferenceIterator) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_reference_iterator_free(v.ptr)
}

// ReferenceNameIsValid returns whether the reference name is well-formed.
//
// Valid reference names must follow one of two patterns:
//
// 1. Top-level names must contain only capital letters and underscores,
// and must begin and end with a letter. (e.g. "HEAD", "ORIG_HEAD").
//
// 2. Names prefixed with "refs/" can be almost anything. You must avoid
// the characters '~', '^', ':', ' \ ', '?', '[', and '*', and the sequences
// ".." and " @ {" which have special meaning to revparse.
func ReferenceNameIsValid(name string) (bool, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var valid C.int
	ret := C.git_reference_name_is_valid(&valid, cname)
	if ret < 0 {
		return false, MakeGitError(ret)
	}
	return valid == 1, nil
}

const (
	// This should match GIT_REFNAME_MAX in src/refs.h
	_refnameMaxLength = C.size_t(1024)
)

type ReferenceFormat uint

const (
	ReferenceFormatNormal           ReferenceFormat = C.GIT_REFERENCE_FORMAT_NORMAL
	ReferenceFormatAllowOnelevel    ReferenceFormat = C.GIT_REFERENCE_FORMAT_ALLOW_ONELEVEL
	ReferenceFormatRefspecPattern   ReferenceFormat = C.GIT_REFERENCE_FORMAT_REFSPEC_PATTERN
	ReferenceFormatRefspecShorthand ReferenceFormat = C.GIT_REFERENCE_FORMAT_REFSPEC_SHORTHAND
)

// ReferenceNormalizeName normalizes the reference name and checks validity.
//
// This will normalize the reference name by removing any leading slash '/'
// characters and collapsing runs of adjacent slashes between name components
// into a single slash.
//
// See git_reference_symbolic_create() for rules about valid names.
func ReferenceNormalizeName(name string, flags ReferenceFormat) (string, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	buf := (*C.char)(C.malloc(_refnameMaxLength))
	defer C.free(unsafe.Pointer(buf))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_reference_normalize_name(buf, _refnameMaxLength, cname, C.uint(flags))
	if ecode < 0 {
		return "", MakeGitError(ecode)
	}

	return C.GoString(buf), nil
}
