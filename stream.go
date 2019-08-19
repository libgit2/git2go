package git

/*
#include <git2.h>
#include <git2/sys/stream.h>

typedef struct {
	git_stream parent;
	void *ptr;
} managed_stream;

extern int _go_git_register_tls(void);
extern void _go_git_setup_stream(managed_stream* s, int encrypted, int proxy_support, void *ptr);

*/
import "C"

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"unsafe"
)

// Network stream for libgit2 to use
type Stream interface {
	Encrypted() bool
	ProxySupport() bool
	Connect() error
	Certificate() (Certificate, error)
	SetProxy(ProxyOptions) error
	io.ReadWriteCloser
}

type ManagedStream struct {
	host string
	port string
	conn *tls.Conn
}

func (self *ManagedStream) Encrypted() bool {
	return true
}

func (self *ManagedStream) ProxySupport() bool {
	return false
}

func (self *ManagedStream) Connect() error {
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", self.host, self.port), nil)
	if err != nil {
		return err
	}

	self.conn = conn
	return nil
}

func (self *ManagedStream) Certificate() (Certificate, error) {
	connState := self.conn.ConnectionState()
	cert := Certificate{
		Kind: CertificateX509,
		X509: connState.PeerCertificates[0],
	}

	return cert, nil
}

func (self *ManagedStream) SetProxy(opts ProxyOptions) error {
	return errors.New("proxy not supported")
}

func (self *ManagedStream) Read(p []byte) (int, error) {
	return self.conn.Read(p)
}

func (self *ManagedStream) Write(p []byte) (int, error) {
	return self.conn.Write(p)
}

func (self *ManagedStream) Close() error {
	return self.conn.Close()
}

var errNotStream = errors.New("passed object does not implement Stream")

// getStreamInterface extracts the Stream interface from the pointers we passed
// to the C code.
func getStreamInterface(_s *C.git_stream) (Stream, error) {
	// For type compatibility we accept C.git_stream but we know we pass
	// C.managed_stream so force the casting to that.
	wrapperPtr := (*C.managed_stream)(unsafe.Pointer(_s))

	// Inside we've stored a handle to the actual type, which must implement
	// Stream.
	stream, ok := pointerHandles.Get(wrapperPtr.ptr).(Stream)
	if !ok {
		return nil, errNotStream
	}

	return stream, nil
}

//export streamCertificate
func streamCertificate(out **C.git_cert, _s *C.git_stream) C.int {
	stream, err := getStreamInterface(_s)
	if err != nil {
		return setLibgit2Error(err)
	}

	cert, err := stream.Certificate()
	if err != nil {
		return setLibgit2Error(err)
	}

	ccert, err := cert.toC()
	if err != nil {
		return setLibgit2Error(err)
	}

	*out = ccert
	return 0
}

//export streamSetProxy
func streamSetProxy(s *C.git_stream, proxy_opts *C.git_proxy_options) C.int {
	setLibgit2Error(errors.New("proxy not supported"))
	return -1
}

//export streamConnect
func streamConnect(_s *C.git_stream) C.int {
	stream, err := getStreamInterface(_s)
	if err != nil {
		return setLibgit2Error(err)
	}

	err = stream.Connect()
	if err != nil {
		return setLibgit2Error(err)
	}

	return 0
}

//export streamRead
func streamRead(_s *C.git_stream, data unsafe.Pointer, l C.size_t) C.ssize_t {
	stream, err := getStreamInterface(_s)
	if err != nil {
		setLibgit2Error(err)
		return -1
	}

	var p []byte
	header := (*reflect.SliceHeader)(unsafe.Pointer(&p))
	header.Cap = int(l)
	header.Len = int(l)
	header.Data = uintptr(data)

	n, err := stream.Read(p)
	if err != nil {
		setLibgit2Error(err)
		return -1
	}

	return C.ssize_t(n)
}

//export streamWrite
func streamWrite(_s *C.git_stream, data unsafe.Pointer, l C.size_t, _f C.int) C.ssize_t {
	stream, err := getStreamInterface(_s)
	if err != nil {
		setLibgit2Error(err)
		return -1
	}

	var p []byte
	header := (*reflect.SliceHeader)(unsafe.Pointer(&p))
	header.Cap = int(l)
	header.Len = int(l)
	header.Data = uintptr(data)

	n, err := stream.Write(p)
	if err != nil {
		setLibgit2Error(err)
		return -1
	}

	return C.ssize_t(n)
}

//export streamClose
func streamClose(_s *C.git_stream) C.int {
	stream, err := getStreamInterface(_s)
	if err != nil {
		return setLibgit2Error(err)
	}

	err = stream.Close()
	if err != nil {
		return setLibgit2Error(err)
	}

	return 0
}

//export streamFree
func streamFree(_s *C.git_stream) {
	wrapperPtr := (*C.managed_stream)(unsafe.Pointer(_s))
	pointerHandles.Untrack(wrapperPtr.ptr)
}

func newManagedStream(host, port string) *ManagedStream {
	return &ManagedStream{
		host: host,
		port: port,
	}
}

//export streamCallbackCb
func streamCallbackCb(out **C.git_stream, chost, cport *C.char) C.int {
	stream := C.calloc(1, C.size_t(unsafe.Sizeof(C.managed_stream{})))
	managed := newManagedStream(C.GoString(chost), C.GoString(cport))
	managedPtr := pointerHandles.Track(managed)
	C._go_git_setup_stream(stream, 1, 0, managedPtr)

	*out = (*C.git_stream)(stream)
	return 0
}

func setLibgit2Error(err error) C.int {
	cstr := C.CString(err.Error())
	defer C.free(unsafe.Pointer(cstr))
	C.giterr_set_str(C.GITERR_NET, cstr)

	return -1
}

func RegisterManagedTls() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C._go_git_register_tls(); err != 0 {
		return MakeGitError(err)
	}

	return nil
}
