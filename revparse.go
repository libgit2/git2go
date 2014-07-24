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
	From Object
	To   Object
	repo *Repository
}

func newRevSpecFrom(ptr *C.git_revspec, repo *Repository) *RevSpec {
	rev := &RevSpec{
		ptr:  ptr,
		From: allocObject(ptr.from, repo),
		To:   allocObject(ptr.to, repo),
		repo: repo,
	}
	runtime.SetFinalizer(rev, (*RevSpec).Free)

	return rev
}

func (r *RevSpec) Free() {
	runtime.SetFinalizer(r, nil)
	r.From.Free()
	r.To.Free()
}

func (r *Repository) RevParse(spec string) (*RevSpec, error) {
	cspec := C.CString(spec)
	defer C.free(unsafe.Pointer(cspec))
	var ptr *C.git_revspec

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_revparse(ptr, r.ptr, cspec)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newRevSpecFrom(ptr, r), nil
}
