package git

/*
#include <git2.h>
#include <git2/errors.h>

extern void _go_git_revspec_free(git_revspec *revspec);
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type RevSpec struct {
	ptr  *C.git_revspec
	repo *Repository
}

func newRevSpecFrom(ptr *C.git_revspec, repo *Repository) *RevSpec {
	rev := &RevSpec{
		ptr:  ptr,
		repo: repo,
	}

	return rev
}

func (r *RevSpec) From() Object {
	if r.ptr.from == nil {
		return nil
	}

	return allocObject(r.ptr.from, r.repo)
}

func (r *RevSpec) To() Object {
	if r.ptr.to == nil {
		return nil
	}

	return allocObject(r.ptr.to, r.repo)
}

func (r *Repository) RevParse(spec string) (*RevSpec, error) {
	cspec := C.CString(spec)
	defer C.free(unsafe.Pointer(cspec))

	ptr := new(C.git_revspec)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_revparse(ptr, r.ptr, cspec)
	if ecode != 0 {
		return nil, MakeGitError(ecode)
	}

	return newRevSpecFrom(ptr, r), nil
}

func (r *Repository) RevParseSingle(spec string) (Object, error) {
	cspec := C.CString(spec)
	defer C.free(unsafe.Pointer(cspec))

	var obj *C.git_object

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_revparse_single(&obj, r.ptr, cspec)
	if ecode != 0 {
		return nil, MakeGitError(ecode)
	}

	return allocObject(obj, r), nil
}
