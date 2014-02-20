package git

/*
#include <git2.h>
*/
import "C"
import (
	"runtime"
)

type Patch struct {
	ptr *C.git_patch
}

func newPatch(ptr *C.git_patch) *Patch {
	if ptr == nil {
		return nil
	}

	patch := &Patch{
		ptr: ptr,
	}

	runtime.SetFinalizer(patch, (*Patch).Free)
	return patch
}

func (patch *Patch) Free() {
	runtime.SetFinalizer(patch, nil)
	C.git_patch_free(patch.ptr)
}

func (patch *Patch) String() string {
	var cptr *C.char
	C.git_patch_to_str(&cptr, patch.ptr)
	return C.GoString(cptr)
}
