package git

/*
#include <git2.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type Patch struct {
	doNotCompare
	ptr *C.git_patch
}

func newPatchFromC(ptr *C.git_patch) *Patch {
	if ptr == nil {
		return nil
	}

	patch := &Patch{
		ptr: ptr,
	}

	runtime.SetFinalizer(patch, (*Patch).Free)
	return patch
}

func (patch *Patch) Free() error {
	if patch.ptr == nil {
		return ErrInvalid
	}
	runtime.SetFinalizer(patch, nil)
	C.git_patch_free(patch.ptr)
	patch.ptr = nil
	return nil
}

func (patch *Patch) String() (string, error) {
	if patch.ptr == nil {
		return "", ErrInvalid
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var buf C.git_buf

	ecode := C.git_patch_to_buf(&buf, patch.ptr)
	runtime.KeepAlive(patch)
	if ecode < 0 {
		return "", MakeGitError(ecode)
	}
	defer C.git_buf_dispose(&buf)

	return C.GoString(buf.ptr), nil
}

func toPointer(data []byte) (ptr unsafe.Pointer) {
	if len(data) > 0 {
		ptr = unsafe.Pointer(&data[0])
	} else {
		ptr = unsafe.Pointer(nil)
	}
	return
}

func (v *Repository) PatchFromBuffers(oldPath, newPath string, oldBuf, newBuf []byte, opts *DiffOptions) (*Patch, error) {
	var patchPtr *C.git_patch

	oldPtr := toPointer(oldBuf)
	newPtr := toPointer(newBuf)

	cOldPath := C.CString(oldPath)
	defer C.free(unsafe.Pointer(cOldPath))

	cNewPath := C.CString(newPath)
	defer C.free(unsafe.Pointer(cNewPath))

	var err error
	copts := populateDiffOptions(&C.git_diff_options{}, opts, v, &err)
	defer freeDiffOptions(copts)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_patch_from_buffers(&patchPtr, oldPtr, C.size_t(len(oldBuf)), cOldPath, newPtr, C.size_t(len(newBuf)), cNewPath, copts)
	runtime.KeepAlive(oldBuf)
	runtime.KeepAlive(newBuf)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return nil, err
	}
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newPatchFromC(patchPtr), nil
}
