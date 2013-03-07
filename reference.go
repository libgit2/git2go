package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

var (
	SYMBOLIC = C.GIT_REF_SYMBOLIC
	OID      = C.GIT_REF_OID
)

type Reference struct {
	ptr *C.git_reference
}

func newReferenceFromC(ptr *C.git_reference) *Reference {
	ref := &Reference{ptr}
	runtime.SetFinalizer(ref, (*Reference).Free)

	return ref
}

func (v *Reference) SetSymbolicTarget(target string) (*Reference, error) {
	var ptr *C.git_reference
	ctarget := C.CString(target)
	defer C.free(unsafe.Pointer(ctarget))

	ret := C.git_reference_symbolic_set_target(&ptr, v.ptr, ctarget)
	if ret < 0 {
		return nil, LastError()
	}

	return newReferenceFromC(ptr), nil
}

func (v *Reference) SetTarget(target *Oid) (*Reference, error) {
	var ptr *C.git_reference

	ret := C.git_reference_set_target(&ptr, v.ptr, target.toC())
	if ret < 0 {
		return nil, LastError()
	}

	return newReferenceFromC(ptr), nil
}

func (v *Reference) Resolve() (*Reference, error) {
	var ptr *C.git_reference

	ret := C.git_reference_resolve(&ptr, v.ptr)
	if ret < 0 {
		return nil, LastError()
	}

	return newReferenceFromC(ptr), nil
}

func (v *Reference) Rename(name string, force bool) (*Reference, error) {
	var ptr *C.git_reference
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ret := C.git_reference_rename(&ptr, v.ptr, cname, cbool(force))

	if ret < 0 {
		return nil, LastError()
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
	ret := C.git_reference_delete(v.ptr)

	if ret < 0 {
		return LastError()
	}

	return nil
}

func (v *Reference) Name() string {
	return C.GoString(C.git_reference_name(v.ptr))
}

func (v *Reference) Type() int {
	return int(C.git_reference_type(v.ptr))
}

func (v *Reference) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_reference_free(v.ptr)
}
