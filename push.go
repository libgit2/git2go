package git

/*
#include <git2.h>
#include <git2/errors.h>

int _go_git_push_status_foreach(git_push *push, void *data);
int _go_git_push_set_callbacks(git_push *push, void *packbuilder_progress_data, void *transfer_progress_data);

*/
import "C"
import (
	"runtime"
	"unsafe"
)

type Push struct {
	ptr *C.git_push

	packbuilderProgress *PackbuilderProgressCallback
	transferProgress    *PushTransferProgressCallback
}

func newPushFromC(cpush *C.git_push) *Push {
	p := &Push{ptr: cpush}
	runtime.SetFinalizer(p, (*Push).Free)
	return p
}

func (p *Push) Free() {
	runtime.SetFinalizer(p, nil)
	C.git_push_free(p.ptr)
}

func (remote *Remote) NewPush() (*Push, error) {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var cpush *C.git_push
	ret := C.git_push_new(&cpush, remote.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return newPushFromC(cpush), nil
}

func (p *Push) Finish() error {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_push_finish(p.ptr)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (p *Push) UnpackOk() bool {

	ret := C.git_push_unpack_ok(p.ptr)
	if ret == 0 {
		return false
	}
	return true

}

func (p *Push) UpdateTips(sig *Signature, msg string) error {

	var csig *C.git_signature = nil
	if sig != nil {
		csig = sig.toC()
		defer C.free(unsafe.Pointer(csig))
	}

	var cmsg *C.char
	if msg == "" {
		cmsg = nil
	} else {
		cmsg = C.CString(msg)
		defer C.free(unsafe.Pointer(cmsg))
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_push_update_tips(p.ptr, csig, cmsg)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (p *Push) AddRefspec(refspec string) error {

	crefspec := C.CString(refspec)
	defer C.free(unsafe.Pointer(crefspec))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_push_add_refspec(p.ptr, crefspec)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

type PushOptions struct {
	Version       uint
	PbParallelism uint
}

func (p *Push) SetOptions(opts PushOptions) error {
	copts := C.git_push_options{version: C.uint(opts.Version), pb_parallelism: C.uint(opts.PbParallelism)}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_push_set_options(p.ptr, &copts)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

type StatusForeachFunc func(ref string, msg string) int

//export statusForeach
func statusForeach(_ref *C.char, _msg *C.char, _data unsafe.Pointer) C.int {
	ref := C.GoString(_ref)
	msg := C.GoString(_msg)

	cb := (*StatusForeachFunc)(_data)

	return C.int((*cb)(ref, msg))
}

func (p *Push) StatusForeach(callback StatusForeachFunc) error {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C._go_git_push_status_foreach(p.ptr, unsafe.Pointer(&callback))
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil

}

type PushCallbacks struct {
	PackbuilderProgress *PackbuilderProgressCallback
	TransferProgress    *PushTransferProgressCallback
}

type PackbuilderProgressCallback func(stage int, current uint, total uint) int
type PushTransferProgressCallback func(current uint, total uint, bytes uint) int

//export packbuilderProgress
func packbuilderProgress(stage C.int, current C.uint, total C.uint, data unsafe.Pointer) C.int {
	return C.int((*(*PackbuilderProgressCallback)(data))(int(stage), uint(current), uint(total)))
}

//export pushTransferProgress
func pushTransferProgress(current C.uint, total C.uint, bytes C.size_t, data unsafe.Pointer) C.int {
	return C.int((*(*PushTransferProgressCallback)(data))(uint(current), uint(total), uint(bytes)))
}

func (p *Push) SetCallbacks(callbacks PushCallbacks) {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// save callbacks so they don't get GC'd
	p.packbuilderProgress = callbacks.PackbuilderProgress
	p.transferProgress = callbacks.TransferProgress

	C._go_git_push_set_callbacks(p.ptr, unsafe.Pointer(p.packbuilderProgress), unsafe.Pointer(p.transferProgress))
}
