package git

/*
#include <string.h>

#include <git2.h>
#include <git2/sys/transport.h>

typedef struct {
	git_smart_subtransport parent;
	void *handle;
} _go_managed_smart_subtransport;

typedef struct {
	git_smart_subtransport_stream parent;
	void *handle;
} _go_managed_smart_subtransport_stream;

int _go_git_transport_register(const char *prefix, void *handle);
int _go_git_transport_smart(git_transport **out, git_remote *owner, int stateless, _go_managed_smart_subtransport *subtransport_payload);
void _go_git_setup_smart_subtransport_stream(_go_managed_smart_subtransport_stream *stream);
*/
import "C"
import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

var (
	// globalRegisteredSmartTransports is a mapping of global, git2go-managed
	// transports.
	globalRegisteredSmartTransports = struct {
		sync.Mutex
		transports map[string]*RegisteredSmartTransport
	}{
		transports: make(map[string]*RegisteredSmartTransport),
	}
)

// unregisterManagedTransports unregisters all git2go-managed transports.
func unregisterManagedTransports() error {
	globalRegisteredSmartTransports.Lock()
	originalTransports := globalRegisteredSmartTransports.transports
	globalRegisteredSmartTransports.transports = make(map[string]*RegisteredSmartTransport)
	globalRegisteredSmartTransports.Unlock()

	var err error
	for protocol, managed := range originalTransports {
		unregisterErr := managed.Free()
		if err == nil && unregisterErr != nil {
			err = fmt.Errorf("failed to unregister transport for %q: %v", protocol, unregisterErr)
		}
	}
	return err
}

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

// Transport encapsulates a way to communicate with a Remote.
type Transport struct {
	doNotCompare
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
func (t *Transport) SmartCredentials(user string, methods CredentialType) (*Credential, error) {
	cred := newCredential()
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

// SmartCertificateCheck calls the certificate check for this transport.
func (t *Transport) SmartCertificateCheck(cert *Certificate, valid bool, hostname string) error {
	var ccert *C.git_cert
	switch cert.Kind {
	case CertificateHostkey:
		chostkeyCert := C.git_cert_hostkey{
			parent: C.git_cert{
				cert_type: C.GIT_CERT_HOSTKEY_LIBSSH2,
			},
			_type:       C.git_cert_ssh_t(cert.Kind),
			hostkey:     (*C.char)(C.CBytes(cert.Hostkey.Hostkey)),
			hostkey_len: C.size_t(len(cert.Hostkey.Hostkey)),
		}
		defer C.free(unsafe.Pointer(chostkeyCert.hostkey))
		C.memcpy(unsafe.Pointer(&chostkeyCert.hash_md5[0]), unsafe.Pointer(&cert.Hostkey.HashMD5[0]), C.size_t(len(cert.Hostkey.HashMD5)))
		C.memcpy(unsafe.Pointer(&chostkeyCert.hash_sha1[0]), unsafe.Pointer(&cert.Hostkey.HashSHA1[0]), C.size_t(len(cert.Hostkey.HashSHA1)))
		C.memcpy(unsafe.Pointer(&chostkeyCert.hash_sha256[0]), unsafe.Pointer(&cert.Hostkey.HashSHA256[0]), C.size_t(len(cert.Hostkey.HashSHA256)))
		if cert.Hostkey.SSHPublicKey.Type() == "ssh-rsa" {
			chostkeyCert.raw_type = C.GIT_CERT_SSH_RAW_TYPE_RSA
		} else if cert.Hostkey.SSHPublicKey.Type() == "ssh-dss" {
			chostkeyCert.raw_type = C.GIT_CERT_SSH_RAW_TYPE_DSS
		} else {
			chostkeyCert.raw_type = C.GIT_CERT_SSH_RAW_TYPE_UNKNOWN
		}
		ccert = (*C.git_cert)(unsafe.Pointer(&chostkeyCert))

	case CertificateX509:
		cx509Cert := C.git_cert_x509{
			parent: C.git_cert{
				cert_type: C.GIT_CERT_X509,
			},
			len:  C.size_t(len(cert.X509.Raw)),
			data: C.CBytes(cert.X509.Raw),
		}
		defer C.free(cx509Cert.data)
		ccert = (*C.git_cert)(unsafe.Pointer(&cx509Cert))
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	chostname := C.CString(hostname)
	defer C.free(unsafe.Pointer(chostname))

	cvalid := C.int(0)
	if valid {
		cvalid = C.int(1)
	}

	ret := C.git_transport_smart_certificate_check(t.ptr, ccert, cvalid, chostname)
	if ret != 0 {
		return MakeGitError(ret)
	}

	return nil
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

// RegisteredSmartTransport represents a transport that has been registered.
type RegisteredSmartTransport struct {
	doNotCompare
	name      string
	stateless bool
	callback  SmartSubtransportCallback
	handle    unsafe.Pointer
}

// NewRegisteredSmartTransport adds a custom transport definition, to be used
// in addition to the built-in set of transports that come with libgit2.
func NewRegisteredSmartTransport(
	name string,
	stateless bool,
	callback SmartSubtransportCallback,
) (*RegisteredSmartTransport, error) {
	return newRegisteredSmartTransport(name, stateless, callback, false)
}

func newRegisteredSmartTransport(
	name string,
	stateless bool,
	callback SmartSubtransportCallback,
	global bool,
) (*RegisteredSmartTransport, error) {
	if !global {
		// Check if we had already registered a smart transport for this protocol. If
		// we had, free it. The user is now responsible for this transport for the
		// lifetime of the library.
		globalRegisteredSmartTransports.Lock()
		if managed, ok := globalRegisteredSmartTransports.transports[name]; ok {
			delete(globalRegisteredSmartTransports.transports, name)
			globalRegisteredSmartTransports.Unlock()

			err := managed.Free()
			if err != nil {
				return nil, err
			}
		} else {
			globalRegisteredSmartTransports.Unlock()
		}
	}

	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	registeredSmartTransport := &RegisteredSmartTransport{
		name:      name,
		stateless: stateless,
		callback:  callback,
	}
	registeredSmartTransport.handle = pointerHandles.Track(registeredSmartTransport)

	ret := C._go_git_transport_register(cstr, registeredSmartTransport.handle)
	if ret != 0 {
		pointerHandles.Untrack(registeredSmartTransport.handle)
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(registeredSmartTransport, (*RegisteredSmartTransport).Free)
	return registeredSmartTransport, nil
}

// Free releases all resources used by the RegisteredSmartTransport and
// unregisters the custom transport definition referenced by it.
func (t *RegisteredSmartTransport) Free() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cstr := C.CString(t.name)
	defer C.free(unsafe.Pointer(cstr))

	if ret := C.git_transport_unregister(cstr); ret < 0 {
		return MakeGitError(ret)
	}

	pointerHandles.Untrack(t.handle)
	runtime.SetFinalizer(t, nil)
	t.handle = nil
	return nil
}

//export smartTransportCallback
func smartTransportCallback(
	errorMessage **C.char,
	out **C.git_transport,
	owner *C.git_remote,
	handle unsafe.Pointer,
) C.int {
	registeredSmartTransport := pointerHandles.Get(handle).(*RegisteredSmartTransport)
	remote, ok := remotePointers.get(owner)
	if !ok {
		err := errors.New("remote pointer not found")
		return setCallbackError(errorMessage, err)
	}

	managed := &managedSmartSubtransport{
		remote:       remote,
		callback:     registeredSmartTransport.callback,
		subtransport: (*C._go_managed_smart_subtransport)(C.calloc(1, C.size_t(unsafe.Sizeof(C._go_managed_smart_subtransport{})))),
	}
	managedHandle := pointerHandles.Track(managed)
	managed.handle = managedHandle
	managed.subtransport.handle = managedHandle

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C._go_git_transport_smart(out, owner, cbool(registeredSmartTransport.stateless), managed.subtransport)
	if ret != 0 {
		pointerHandles.Untrack(managedHandle)
	}
	return ret
}

//export smartTransportSubtransportCallback
func smartTransportSubtransportCallback(
	errorMessage **C.char,
	wrapperPtr *C._go_managed_smart_subtransport,
	owner *C.git_transport,
) C.int {
	subtransport := pointerHandles.Get(wrapperPtr.handle).(*managedSmartSubtransport)

	underlyingSmartSubtransport, err := subtransport.callback(subtransport.remote, &Transport{ptr: owner})
	if err != nil {
		return setCallbackError(errorMessage, err)
	}
	subtransport.underlying = underlyingSmartSubtransport
	return C.int(ErrorCodeOK)
}

type managedSmartSubtransport struct {
	owner                *C.git_transport
	callback             SmartSubtransportCallback
	remote               *Remote
	subtransport         *C._go_managed_smart_subtransport
	underlying           SmartSubtransport
	handle               unsafe.Pointer
	currentManagedStream *managedSmartSubtransportStream
}

func getSmartSubtransportInterface(subtransport *C.git_smart_subtransport) *managedSmartSubtransport {
	wrapperPtr := (*C._go_managed_smart_subtransport)(unsafe.Pointer(subtransport))
	return pointerHandles.Get(wrapperPtr.handle).(*managedSmartSubtransport)
}

//export smartSubtransportActionCallback
func smartSubtransportActionCallback(
	errorMessage **C.char,
	out **C.git_smart_subtransport_stream,
	t *C.git_smart_subtransport,
	url *C.char,
	action C.git_smart_service_t,
) C.int {
	subtransport := getSmartSubtransportInterface(t)

	underlyingStream, err := subtransport.underlying.Action(C.GoString(url), SmartServiceAction(action))
	if err != nil {
		return setCallbackError(errorMessage, err)
	}

	// It's okay to do strict equality here: we expect both to be identical.
	if subtransport.currentManagedStream == nil || subtransport.currentManagedStream.underlying != underlyingStream {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		stream := (*C._go_managed_smart_subtransport_stream)(C.calloc(1, C.size_t(unsafe.Sizeof(C._go_managed_smart_subtransport_stream{}))))
		managed := &managedSmartSubtransportStream{
			underlying: underlyingStream,
			streamPtr:  stream,
		}
		managedHandle := pointerHandles.Track(managed)
		managed.handle = managedHandle
		stream.handle = managedHandle

		C._go_git_setup_smart_subtransport_stream(stream)

		subtransport.currentManagedStream = managed
	}

	*out = &subtransport.currentManagedStream.streamPtr.parent
	return C.int(ErrorCodeOK)
}

//export smartSubtransportCloseCallback
func smartSubtransportCloseCallback(errorMessage **C.char, t *C.git_smart_subtransport) C.int {
	subtransport := getSmartSubtransportInterface(t)

	subtransport.currentManagedStream = nil

	if subtransport.underlying != nil {
		err := subtransport.underlying.Close()
		if err != nil {
			return setCallbackError(errorMessage, err)
		}
	}

	return C.int(ErrorCodeOK)
}

//export smartSubtransportFreeCallback
func smartSubtransportFreeCallback(t *C.git_smart_subtransport) {
	subtransport := getSmartSubtransportInterface(t)

	if subtransport.underlying != nil {
		subtransport.underlying.Free()
		subtransport.underlying = nil
	}
	pointerHandles.Untrack(subtransport.handle)
	C.free(unsafe.Pointer(subtransport.subtransport))
	subtransport.handle = nil
	subtransport.subtransport = nil
}

type managedSmartSubtransportStream struct {
	owner      *C.git_smart_subtransport_stream
	streamPtr  *C._go_managed_smart_subtransport_stream
	underlying SmartSubtransportStream
	handle     unsafe.Pointer
}

func getSmartSubtransportStreamInterface(subtransportStream *C.git_smart_subtransport_stream) *managedSmartSubtransportStream {
	managedSubtransportStream := (*C._go_managed_smart_subtransport_stream)(unsafe.Pointer(subtransportStream))
	return pointerHandles.Get(managedSubtransportStream.handle).(*managedSmartSubtransportStream)
}

//export smartSubtransportStreamReadCallback
func smartSubtransportStreamReadCallback(
	errorMessage **C.char,
	s *C.git_smart_subtransport_stream,
	buffer *C.char,
	bufSize C.size_t,
	bytesRead *C.size_t,
) C.int {
	stream := getSmartSubtransportStreamInterface(s)

	var p []byte
	header := (*reflect.SliceHeader)(unsafe.Pointer(&p))
	header.Cap = int(bufSize)
	header.Len = int(bufSize)
	header.Data = uintptr(unsafe.Pointer(buffer))

	n, err := stream.underlying.Read(p)
	*bytesRead = C.size_t(n)
	if n == 0 && err != nil {
		if err == io.EOF {
			return C.int(ErrorCodeOK)
		}

		return setCallbackError(errorMessage, err)
	}

	return C.int(ErrorCodeOK)
}

//export smartSubtransportStreamWriteCallback
func smartSubtransportStreamWriteCallback(
	errorMessage **C.char,
	s *C.git_smart_subtransport_stream,
	buffer *C.char,
	bufLen C.size_t,
) C.int {
	stream := getSmartSubtransportStreamInterface(s)

	var p []byte
	header := (*reflect.SliceHeader)(unsafe.Pointer(&p))
	header.Cap = int(bufLen)
	header.Len = int(bufLen)
	header.Data = uintptr(unsafe.Pointer(buffer))

	if _, err := stream.underlying.Write(p); err != nil {
		return setCallbackError(errorMessage, err)
	}

	return C.int(ErrorCodeOK)
}

//export smartSubtransportStreamFreeCallback
func smartSubtransportStreamFreeCallback(s *C.git_smart_subtransport_stream) {
	stream := getSmartSubtransportStreamInterface(s)

	stream.underlying.Free()
	pointerHandles.Untrack(stream.handle)
	C.free(unsafe.Pointer(stream.streamPtr))
	stream.handle = nil
	stream.streamPtr = nil
}
