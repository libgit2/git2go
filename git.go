package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
*/
import "C"
import (
	"unsafe"
)

const (
	ITEROVER  = C.GIT_ITEROVER
	EEXISTS   = C.GIT_EEXISTS
	ENOTFOUND = C.GIT_ENOTFOUND
)

func init() {
	C.git_threads_init()
}

// Oid
type Oid struct {
	bytes [20]byte
}

func newOidFromC(coid *C.git_oid) *Oid {
	oid := new(Oid)
	copy(oid.bytes[0:20], C.GoBytes(unsafe.Pointer(coid), 20))
	return oid
}

func NewOid(b []byte) *Oid {
	oid := new(Oid)
	copy(oid.bytes[0:20], b[0:20])
	return oid
}

func (oid *Oid) toC() *C.git_oid {
	return (*C.git_oid)(unsafe.Pointer(&oid.bytes))
}

func NewOidFromString(s string) (*Oid, error) {
	o := new(Oid)
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))

	if C.git_oid_fromstr(o.toC(), cs) < 0 {
		return nil, LastError()
	}

	return o, nil
}

func (oid *Oid) String() string {
	buf := make([]byte, 40)
	C.git_oid_fmt((*C.char)(unsafe.Pointer(&buf[0])), oid.toC())
	return string(buf)
}

func (oid *Oid) Bytes() []byte {
	return oid.bytes[0:]
}

type GitError struct {
	Message string
	Code int
}

func (e GitError) Error() string{
	return e.Message
}

func LastError() error {
	err := C.giterr_last()
	return &GitError{C.GoString(err.message), int(err.klass)}
}

func cbool(b bool) C.int {
	if (b) {
		return C.int(1)
	}
	return C.int(0)
}

func ucbool(b bool) C.uint {
	if (b) {
		return C.uint(1)
	}
	return C.uint(0)
}
