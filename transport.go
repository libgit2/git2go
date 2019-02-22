package git

/*
#include <git2.h>
#include <git2/sys/transport.h>

extern int _go_git_transport_register(const char *scheme, void *ptr);
extern void _go_git_init_transport_interface_wrapper(git_transport *transport);
extern int _go_git_call_progress_cb(git_transfer_progress_cb progress_cb, git_transfer_progress *stats, void *progress_payload);
extern void _go_git_call_push_transfer_progress(git_push_transfer_progress cb, unsigned int written, unsigned int total, size_t bytes, void *payload);

*/
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// int git_transport_cb(git_transport **out, git_remote *owner, void *param);
type CreateTransportCallback func(owner *Remote) (Transport, error)

//export CallbackGitCreateTransport
func CallbackGitCreateTransport(transportPtr **C.git_transport, remotePtr *C.git_remote, cbPtr unsafe.Pointer) C.int {
	p := pointerHandles.Get(cbPtr)
	if callback, ok := p.(CreateTransportCallback); ok {
		remote := &Remote{ptr: remotePtr}

		goTransport, err := callback(remote)
		if err != nil {
			return -1
		}

		cTransport := &C.git_transport{}
		ret := C.git_transport_init(cTransport, C.GIT_TRANSPORT_VERSION)
		if ret < 0 {
			return ret
		}

		C._go_git_init_transport_interface_wrapper(cTransport)

		transportHandles.Track(cTransport, goTransport)
		*transportPtr = cTransport
		return 0

	} else {
		panic("invalid transport callback")
	}
}

func RegisterTransport(scheme string, callback CreateTransportCallback) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cscheme := C.CString(scheme)
	defer C.free(unsafe.Pointer(cscheme))

	cb_ptr := pointerHandles.Track(callback)

	err := C._go_git_transport_register(cscheme, cb_ptr)
	if err < 0 {
		return MakeGitError(err)
	}

	return nil
}

type Transport interface {
	// int (*set_custom_headers)(
	//     git_transport *transport,
	//     const git_strarray *custom_headers);
	SetCustomHeaders(headers []string) error

	// int (*connect)(
	//     git_transport *transport,
	//     const char *url,
	//     git_cred_acquire_cb cred_acquire_cb,
	//     void *cred_acquire_payload,
	//     const git_proxy_options *proxy_opts,
	//     int direction,
	//     int flags);
	Connect(url string) error

	// int (*ls)(
	//     const git_remote_head ***out,
	//     size_t *size,
	//     git_transport *transport);
	Ls() ([]RemoteHead, error)

	// int (*negotiate_fetch)(
	//     git_transport *transport,
	//     git_repository *repo,
	//     const git_remote_head * const *refs,
	//     size_t count);
	NegotiateFetch(r *Repository, wants []RemoteHead) error

	// int (*download_pack)(
	//     git_transport *transport,
	//     git_repository *repo,
	//     git_transfer_progress *stats,
	//     git_transfer_progress_cb progress_cb,
	//     void *progress_payload);
	DownloadPack(r *Repository, progress *TransferProgress, progressCb TransferProgressCallback) error

	// int (*is_connected)(git_transport *transport);
	IsConnected() (bool, error)

	// void (*cancel)(git_transport *transport);
	Cancel()

	// int (*close)(git_transport *transport);
	Close() error

	// void (*free)(git_transport *transport);
	Free()

	// Push is apparently not supposed to be implemented by anything not conforming to the smart
	// protocol (see smart_protocol.c).  Seems strange, because it's clearly part of this interface.
	// For more info: https://github.com/libgit2/libgit2/issues/4648
	//
	// int(*push)(git_transport *transport, git_push *push, const git_remote_callbacks *callbacks);
	// Push(progressCb func(objsWritten, objsTotal uint32, bytes uint)) error
}

//export transportSetCustomHeadersCallback
func transportSetCustomHeadersCallback(cTransport *C.git_transport, cHeaders *C.git_strarray) int {
	goTransport := transportHandles.Get(cTransport)

	err := goTransport.SetCustomHeaders(makeStringsFromCStrings(cHeaders.strings, int(cHeaders.count)))
	if err != nil {
		return -1
	}

	return 0
}

//export transportConnectCallback
func transportConnectCallback(
	cTransport *C.git_transport,
	url *C.char,
	cred_acquire_cb C.git_cred_acquire_cb,
	cred_acquire_payload unsafe.Pointer,
	proxy_opts *C.git_proxy_options,
	direction int,
	flags int,
) int {
	goTransport := transportHandles.Get(cTransport)

	err := goTransport.Connect(C.GoString(url))
	if err != nil {
		return -1
	}

	return 0
}

//export transportLsCallback
func transportLsCallback(outRemoteHeads ***C.git_remote_head, outNumHeads *C.size_t, cTransport *C.git_transport) int {
	goTransport := transportHandles.Get(cTransport)

	goRemoteHeads, err := goTransport.Ls()
	if err != nil {
		return -1
	}

	*outNumHeads = C.size_t(len(goRemoteHeads))
	cRemoteHeads := make([]*C.git_remote_head, len(goRemoteHeads))
	for i := 0; i < len(goRemoteHeads); i++ {
		cRemoteHeads[i] = goRemoteHeads[i].toC()
	}
	if len(cRemoteHeads) > 0 {
		*outRemoteHeads = (**C.git_remote_head)(unsafe.Pointer(&cRemoteHeads[0]))
	}

	return 0
}

//export transportNegotiateFetchCallback
func transportNegotiateFetchCallback(cTransport *C.git_transport, cRepo *C.git_repository, cRefs **C.git_remote_head, numRefs C.size_t) int {
	goTransport := transportHandles.Get(cTransport)

	repo := newRepositoryFromCNoFinalizer(cRepo)

	err := goTransport.NegotiateFetch(repo, newRemoteHeadsFromC(cRefs, numRefs))
	if err != nil {
		return -1
	}
	return 0
}

//export transportDownloadPackCallback
func transportDownloadPackCallback(
	cTransport *C.git_transport,
	cRepo *C.git_repository,
	cProgress *C.git_transfer_progress,
	progress_cb C.git_transfer_progress_cb,
	progress_payload unsafe.Pointer,
) int {
	goTransport := transportHandles.Get(cTransport)

	repo := newRepositoryFromCNoFinalizer(cRepo)
	progress := newTransferProgressFromC(cProgress)

	progressCb := func(stats TransferProgress) ErrorCode {
		return ErrorCode(C._go_git_call_progress_cb(progress_cb, stats.toC(), progress_payload))
	}

	err := goTransport.DownloadPack(repo, &progress, progressCb)
	if err != nil {
		return -1
	}
	return 0
}

//export transportIsConnectedCallback
func transportIsConnectedCallback(cTransport *C.git_transport) int {
	goTransport := transportHandles.Get(cTransport)

	is, err := goTransport.IsConnected()
	if err != nil {
		return -1
	} else if is {
		return 1
	}
	return 0
}

//export transportCancelCallback
func transportCancelCallback(cTransport *C.git_transport) {
	goTransport := transportHandles.Get(cTransport)
	goTransport.Cancel()
}

//export transportCloseCallback
func transportCloseCallback(cTransport *C.git_transport) int {
	goTransport := transportHandles.Get(cTransport)
	err := goTransport.Close()
	if err != nil {
		return -1
	}
	return 0
}

//export transportFreeCallback
func transportFreeCallback(cTransport *C.git_transport) {
	goTransport := transportHandles.Get(cTransport)
	goTransport.Free()

	transportHandles.Untrack(cTransport)
}

// Maps a registered Transport instance to a C git_transport pointer so that Go code can retrieve
// the Transport given only a git_transport*.
type transportHandleList struct {
	sync.RWMutex
	handles map[*C.git_transport]Transport
}

var transportHandles = newTransportHandleList()

func newTransportHandleList() *transportHandleList {
	return &transportHandleList{
		handles: make(map[*C.git_transport]Transport),
	}
}

func (v *transportHandleList) Track(cTransport *C.git_transport, goTransport Transport) {
	v.Lock()
	v.handles[cTransport] = goTransport
	v.Unlock()
}

func (v *transportHandleList) Untrack(cTransport *C.git_transport) {
	v.Lock()
	delete(v.handles, cTransport)
	v.Unlock()
}

func (v *transportHandleList) Get(cTransport *C.git_transport) Transport {
	v.RLock()
	defer v.RUnlock()

	ptr, ok := v.handles[cTransport]
	if !ok {
		panic(fmt.Sprintf("invalid pointer handle: %p", cTransport))
	}

	return ptr
}
