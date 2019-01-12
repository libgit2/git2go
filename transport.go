package git

/*
#include <git2.h>
#include <git2/sys/transport.h>

typedef struct {
	git_smart_subtransport parent;
	void *ptr;
} _go_managed_smart_subtransport;

typedef struct {
	git_smart_subtransport_stream parent;
	void *ptr;
} _go_managed_smart_subtransport_stream;

int _go_git_smart_transport_register(const char *scheme, void *payload);
int _go_git_transport_smart(git_transport **out, git_remote *owner, int stateless, void *subtransport_payload);
void _go_git_setup_smart_subtransport_stream(_go_managed_smart_subtransport_stream *stream);
*/
import "C"
import (
	"errors"
	"io"
	"reflect"
	"runtime"
	"unsafe"
)

// SmartServiceAction is an action that the smart transport can ask a
// subtransport to perform.
type SmartServiceAction int

const (
	// SmartServiceActionUploadpackLs is used upon connecting to a remote, and is
	// used to perform reference discovery prior to performing a pull operation.
	SmartServiceActionUploadpackLs SmartServiceAction = C.GIT_SERVICE_UPLOADPACK_LS

	// SmartServiceActionUploadpack is used when performing a pull operation.
	SmartServiceActionUploadpack SmartServiceAction = C.GIT_SERVICE_UPLOADPACK

	// SmartServiceActionReceivepackLs is used upon connecting to a remote, and is
	// used to perform reference discovery prior to performing a push operation.
	SmartServiceActionReceivepackLs SmartServiceAction = C.GIT_SERVICE_RECEIVEPACK_LS

	// SmartServiceActionReceivepack is used when performing a push operation.
	SmartServiceActionReceivepack SmartServiceAction = C.GIT_SERVICE_RECEIVEPACK
)

// RegisteredSmartTransport represents a transport that has been registered.
type RegisteredSmartTransport struct {
	name      string
	stateless bool
	callback  SmartSubtransportCallback
	handle    unsafe.Pointer
}

// Transport encapsulates a way to communicate with a Remote.
type Transport struct {
	ptr *C.git_transport
}

// SmartProxyOptions gets a copy of the proxy options for this transport.
func (t *Transport) SmartProxyOptions() (*ProxyOptions, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var cpopts C.git_proxy_options
	if ret := C.git_transport_smart_proxy_options(&cpopts, t.ptr); ret < 0 {
		return nil, MakeGitError(ret)
	}

	return proxyOptionsFromC(&cpopts), nil
}

// SmartCredentials calls the credentials callback for this transport.
func (t *Transport) SmartCredentials(user string, methods CredType) (*Cred, error) {
	cred := newCred()
	var cstr *C.char

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if user != "" {
		cstr = C.CString(user)
		defer C.free(unsafe.Pointer(cstr))
	}
	ret := C.git_transport_smart_credentials(&cred.ptr, t.ptr, cstr, C.int(methods))
	if ret != 0 {
		cred.Free()
		return nil, MakeGitError(ret)
	}

	return cred, nil
}

// SmartSubtransport is the interface for custom subtransports which carry data
// for the smart transport.
type SmartSubtransport interface {
	// Action creates a SmartSubtransportStream for the provided url and
	// requested action.
	Action(url string, action SmartServiceAction) (SmartSubtransportStream, error)

	// Close closes the SmartSubtransport.
	//
	// Subtransports are guaranteed a call to Close between
	// calls to Action, except for the following two "natural" progressions
	// of actions against a constant URL.
	//
	// 1. UPLOADPACK_LS -> UPLOADPACK
	// 2. RECEIVEPACK_LS -> RECEIVEPACK
	Close() error

	// Free releases the resources of the SmartSubtransport.
	Free()
}

// SmartSubtransportStream is the interface for streams used by the smart
// transport to read and write data from a subtransport.
type SmartSubtransportStream interface {
	io.Reader
	io.Writer

	// Free releases the resources of the SmartSubtransportStream.
	Free()
}

// SmartSubtransportCallback is a function which creates a new subtransport for
// the smart transport.
type SmartSubtransportCallback func(remote *Remote, transport *Transport) (SmartSubtransport, error)

// NewRegisteredSmartTransport adds a custom transport definition, to be used
// in addition to the built-in set of transports that come with libgit2.
func NewRegisteredSmartTransport(
	name string,
	stateless bool,
	callback SmartSubtransportCallback,
) (*RegisteredSmartTransport, error) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	registeredSmartTransport := new(RegisteredSmartTransport)
	registeredSmartTransport.name = name
	registeredSmartTransport.stateless = stateless
	registeredSmartTransport.callback = callback
	registeredSmartTransport.handle = pointerHandles.Track(registeredSmartTransport)

	ret := C._go_git_smart_transport_register(cstr, registeredSmartTransport.handle)
	runtime.KeepAlive(callback)
	runtime.KeepAlive(cstr)
	if ret != 0 {
		pointerHandles.Untrack(registeredSmartTransport.handle)
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(registeredSmartTransport, (*RegisteredSmartTransport).Free)
	return registeredSmartTransport, nil
}

// Free releases all resources used by the RegisteredSmartTransport and
// unregisters the custom transport definition referenced by it.
func (t *RegisteredSmartTransport) Free() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cstr := C.CString(t.name)
	defer C.free(unsafe.Pointer(cstr))

	C.git_transport_unregister(cstr)

	pointerHandles.Untrack(t.handle)
	runtime.SetFinalizer(t, nil)
	t.handle = nil
}

//export smartTransportCb
func smartTransportCb(out **C.git_transport, owner *C.git_remote, param unsafe.Pointer) C.int {
	if out == nil {
		return -1
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	registeredSmartTransport := pointerHandles.Get(param).(*RegisteredSmartTransport)
	remote, ok := remotePointers.Get(owner)
	if !ok {
		remote = nil
	}

	subtransport := C.calloc(1, C.size_t(unsafe.Sizeof(C._go_managed_smart_subtransport{})))
	managed := &wrappedSmartSubtransport{
		remote:          remote,
		callback:        registeredSmartTransport.callback,
		subtransportPtr: subtransport,
	}
	managedPtr := pointerHandles.Track(managed)
	managed.ptr = managedPtr
	(*C._go_managed_smart_subtransport)(subtransport).ptr = managedPtr

	stateless := C.int(0)
	if registeredSmartTransport.stateless {
		stateless = C.int(1)
	}

	ret := C._go_git_transport_smart(out, owner, stateless, subtransport)
	if ret != 0 {
		pointerHandles.Untrack(managedPtr)
	}
	return ret
}

//export smartTransportSubtransportCb
func smartTransportSubtransportCb(wrapperPtr *C._go_managed_smart_subtransport, owner *C.git_transport) C.int {
	subtransport, ok := pointerHandles.Get(wrapperPtr.ptr).(*wrappedSmartSubtransport)
	if !ok {
		return setLibgit2Error(ErrClassNet, errNoWrappedSmartSubtransport)
	}
	subtransport.transport = &Transport{
		ptr: owner,
	}
	underlyingSmartSubtransport, err := subtransport.callback(subtransport.remote, subtransport.transport)
	if err != nil {
		return setLibgit2Error(ErrClassNet, err)
	}
	subtransport.underlying = underlyingSmartSubtransport
	return 0
}

type wrappedSmartSubtransport struct {
	owner           *C.git_transport
	callback        SmartSubtransportCallback
	ptr             unsafe.Pointer
	remote          *Remote
	transport       *Transport
	subtransportPtr unsafe.Pointer
	underlying      SmartSubtransport
}

var errNoWrappedSmartSubtransport = errors.New("passed object is not a wrappedSmartSubtransport")

func getSmartSubtransportInterface(_t *C.git_smart_subtransport) (*wrappedSmartSubtransport, error) {
	wrapperPtr := (*C._go_managed_smart_subtransport)(unsafe.Pointer(_t))

	subtransport, ok := pointerHandles.Get(wrapperPtr.ptr).(*wrappedSmartSubtransport)
	if !ok {
		return nil, errNoWrappedSmartSubtransport
	}

	return subtransport, nil
}

//export smartSubtransportActionCb
func smartSubtransportActionCb(out **C.git_smart_subtransport_stream, t *C.git_smart_subtransport, url *C.char, action C.git_smart_service_t) C.int {
	subtransport, err := getSmartSubtransportInterface(t)
	if err != nil {
		return setLibgit2Error(ErrClassNet, err)
	}

	underlyingStream, err := subtransport.underlying.Action(C.GoString(url), SmartServiceAction(action))
	if err != nil {
		return setLibgit2Error(ErrClassNet, err)
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	stream := C.calloc(1, C.size_t(unsafe.Sizeof(C._go_managed_smart_subtransport_stream{})))
	managed := &wrappedSmartSubtransportStream{
		underlying: underlyingStream,
		streamPtr:  stream,
	}
	managedPtr := pointerHandles.Track(managed)
	managed.ptr = managedPtr
	(*C._go_managed_smart_subtransport_stream)(stream).ptr = managedPtr

	C._go_git_setup_smart_subtransport_stream((*C._go_managed_smart_subtransport_stream)(stream))

	*out = (*C.git_smart_subtransport_stream)(stream)
	return 0
}

//export smartSubtransportCloseCb
func smartSubtransportCloseCb(t *C.git_smart_subtransport) C.int {
	subtransport, err := getSmartSubtransportInterface(t)
	if err != nil {
		return setLibgit2Error(ErrClassNet, err)
	}

	if subtransport.underlying != nil {
		err = subtransport.underlying.Close()
		if err != nil {
			return setLibgit2Error(ErrClassNet, err)
		}
	}

	return 0
}

//export smartSubtransportFreeCb
func smartSubtransportFreeCb(t *C.git_smart_subtransport) {
	subtransport, err := getSmartSubtransportInterface(t)
	if err != nil {
		panic(err)
	}

	if subtransport.underlying != nil {
		subtransport.underlying.Free()
		subtransport.underlying = nil
	}
	pointerHandles.Untrack(subtransport.ptr)
	C.free(subtransport.subtransportPtr)
	subtransport.ptr = nil
	subtransport.subtransportPtr = nil
}

type wrappedSmartSubtransportStream struct {
	owner      *C.git_smart_subtransport_stream
	ptr        unsafe.Pointer
	streamPtr  unsafe.Pointer
	underlying SmartSubtransportStream
}

var errNoSmartSubtransportStream = errors.New("passed object is not a wrappedSmartSubtransportStream")

func getSmartSubtransportStreamInterface(_s *C.git_smart_subtransport_stream) (*wrappedSmartSubtransportStream, error) {
	wrapperPtr := (*C._go_managed_smart_subtransport_stream)(unsafe.Pointer(_s))

	stream, ok := pointerHandles.Get(wrapperPtr.ptr).(*wrappedSmartSubtransportStream)
	if !ok {
		return nil, errNoSmartSubtransportStream
	}

	return stream, nil
}

//export smartSubtransportStreamWriteCb
func smartSubtransportStreamWriteCb(s *C.git_smart_subtransport_stream, buffer *C.char, bufLen C.size_t) C.int {
	stream, err := getSmartSubtransportStreamInterface(s)
	if err != nil {
		return setLibgit2Error(ErrClassNet, err)
	}

	var p []byte
	header := (*reflect.SliceHeader)(unsafe.Pointer(&p))
	header.Cap = int(bufLen)
	header.Len = int(bufLen)
	header.Data = uintptr(unsafe.Pointer(buffer))

	if _, err := stream.underlying.Write(p); err != nil {
		return setLibgit2Error(ErrClassNet, err)
	}

	return 0
}

//export smartSubtransportStreamReadCb
func smartSubtransportStreamReadCb(s *C.git_smart_subtransport_stream, buffer *C.char, bufSize C.size_t, bytesRead *C.size_t) C.int {
	stream, err := getSmartSubtransportStreamInterface(s)
	if err != nil {
		return setLibgit2Error(ErrClassNet, err)
	}

	var p []byte
	header := (*reflect.SliceHeader)(unsafe.Pointer(&p))
	header.Cap = int(bufSize)
	header.Len = int(bufSize)
	header.Data = uintptr(unsafe.Pointer(buffer))

	n, err := stream.underlying.Read(p)
	*bytesRead = C.size_t(n)
	if n == 0 && err != nil {
		if err == io.EOF {
			return 0
		}

		return setLibgit2Error(ErrClassNet, err)
	}

	return 0
}

//export smartSubtransportStreamFreeCb
func smartSubtransportStreamFreeCb(s *C.git_smart_subtransport_stream) {
	stream, err := getSmartSubtransportStreamInterface(s)
	if err != nil {
		panic(err)
	}

	stream.underlying.Free()
	pointerHandles.Untrack(stream.ptr)
	C.free(stream.streamPtr)
	stream.ptr = nil
	stream.streamPtr = nil
}
