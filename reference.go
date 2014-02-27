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

type ReferenceType int

const (
	ReferenceSymbolic ReferenceType = C.GIT_REF_SYMBOLIC
	ReferenceOid                    = C.GIT_REF_OID
)

type Reference struct {
	ptr *C.git_reference
}

func newReferenceFromC(ptr *C.git_reference) *Reference {
	ref := &Reference{ptr}
	runtime.SetFinalizer(ref, (*Reference).Free)

	return ref
}

func (v *Reference) SetSymbolicTarget(target string, sig *Signature, msg string) (*Reference, error) {
	var ptr *C.git_reference

	ctarget := C.CString(target)
	defer C.free(unsafe.Pointer(ctarget))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	csig := sig.toC()
	defer C.free(unsafe.Pointer(csig))

	var cmsg *C.char
	if msg == "" {
		cmsg = nil
	} else {
		cmsg = C.CString(msg)
		defer C.free(unsafe.Pointer(cmsg))
	}

	ret := C.git_reference_symbolic_set_target(&ptr, v.ptr, ctarget, csig, cmsg)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReferenceFromC(ptr), nil
}

func (v *Reference) SetTarget(target *Oid, sig *Signature, msg string) (*Reference, error) {
	var ptr *C.git_reference

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	csig := sig.toC()
	defer C.free(unsafe.Pointer(csig))

	var cmsg *C.char
	if msg == "" {
		cmsg = nil
	} else {
		cmsg = C.CString(msg)
		defer C.free(unsafe.Pointer(cmsg))
	}

	ret := C.git_reference_set_target(&ptr, v.ptr, target.toC(), csig, cmsg)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReferenceFromC(ptr), nil
}

func (v *Reference) Resolve() (*Reference, error) {
	var ptr *C.git_reference

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_resolve(&ptr, v.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReferenceFromC(ptr), nil
}

func (v *Reference) Rename(name string, force bool, sig *Signature, msg string) (*Reference, error) {
	var ptr *C.git_reference
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

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_rename(&ptr, v.ptr, cname, cbool(force), csig, cmsg)

	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReferenceFromC(ptr), nil
}

func (v *Reference) Target() *Oid {
	return newOidFromC(C.git_reference_target(v.ptr))
}

func (v *Reference) SymbolicTarget() string {
	cstr := C.git_reference_symbolic_target(v.ptr)
	if cstr == nil {
		return ""
	}

	return C.GoString(cstr)
}

func (v *Reference) Delete() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_delete(v.ptr)

	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

func (v *Reference) Name() string {
	return C.GoString(C.git_reference_name(v.ptr))
}

func (v *Reference) Type() ReferenceType {
	return ReferenceType(C.git_reference_type(v.ptr))
}

func (v *Reference) IsBranch() bool {
	return C.git_reference_is_branch(v.ptr) == 1
}

func (v *Reference) IsRemote() bool {
	return C.git_reference_is_remote(v.ptr) == 1
}

func (v *Reference) IsTag() bool {
	return C.git_reference_is_tag(v.ptr) == 1
}

func (v *Reference) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_reference_free(v.ptr)
}

type ReferenceIterator struct {
	ptr  *C.git_reference_iterator
	repo *Repository
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

	iter := &ReferenceIterator{repo: repo, ptr: ptr}
	runtime.SetFinalizer(iter, (*ReferenceIterator).Free)
	return iter, nil
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

	iter := &ReferenceIterator{repo: repo, ptr: ptr}
	runtime.SetFinalizer(iter, (*ReferenceIterator).Free)
	return iter, nil
}

// NextName retrieves the next reference name. If the iteration is over,
// the returned error is git.ErrIterOver
func (v *ReferenceIterator) NextName() (string, error) {
	var ptr *C.char

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reference_next_name(&ptr, v.ptr)
	if ret == ITEROVER {
		return "", ErrIterOver
	}
	if ret < 0 {
		return "", MakeGitError(ret)
	}

	return C.GoString(ptr), nil
}

// Create a channel from the iterator. You can use range on the
// returned channel to iterate over all the references names. The channel
// will be closed in case any error is found.
func (v *ReferenceIterator) NameIter() <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		name, err := v.NextName()
		for err == nil {
			ch <- name
			name, err = v.NextName()
		}
	}()

	return ch
}

// Next retrieves the next reference. If the iterationis over, the
// returned error is git.ErrIterOver
func (v *ReferenceIterator) Next() (*Reference, error) {
	var ptr *C.git_reference
	ret := C.git_reference_next(&ptr, v.ptr)
	if ret == ITEROVER {
		return nil, ErrIterOver
	}
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReferenceFromC(ptr), nil
}

// Create a channel from the iterator. You can use range on the
// returned channel to iterate over all the references names. The channel
// will be closed in case any error is found.
func (v *ReferenceIterator) Iter() <-chan *Reference {
	ch := make(chan *Reference)
	go func() {
		defer close(ch)
		name, err := v.Next()
		for err == nil {
			ch <- name
			name, err = v.Next()
		}
	}()

	return ch
}

// Free the reference iterator
func (v *ReferenceIterator) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_reference_iterator_free(v.ptr)
}
