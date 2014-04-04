package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
*/
import "C"
import (
	"bytes"
	"encoding/hex"
	"errors"
	"runtime"
	"strings"
	"unsafe"
)

const (
	ITEROVER  = C.GIT_ITEROVER
	EEXISTS   = C.GIT_EEXISTS
	ENOTFOUND = C.GIT_ENOTFOUND
)

var (
	ErrIterOver = errors.New("Iteration is over")
	ErrInvalid  = errors.New("Invalid state for operation")
)

func init() {
	C.git_threads_init()
}

// Oid represents the id for a Git object.
type Oid [20]byte

func newOidFromC(coid *C.git_oid) *Oid {
	if coid == nil {
		return nil
	}

	oid := new(Oid)
	copy(oid[0:20], C.GoBytes(unsafe.Pointer(coid), 20))
	return oid
}

func NewOidFromBytes(b []byte) *Oid {
	oid := new(Oid)
	copy(oid[0:20], b[0:20])
	return oid
}

func (oid *Oid) toC() *C.git_oid {
	return (*C.git_oid)(unsafe.Pointer(oid))
}

func NewOid(s string) (*Oid, error) {
	if len(s) > C.GIT_OID_HEXSZ {
		return nil, errors.New("string is too long for oid")
	}

	o := new(Oid)

	slice, error := hex.DecodeString(s)
	if error != nil {
		return nil, error
	}

	copy(o[:], slice[:20])
	return o, nil
}

func (oid *Oid) String() string {
	return hex.EncodeToString(oid[:])
}

func (oid *Oid) Cmp(oid2 *Oid) int {
	return bytes.Compare(oid[:], oid2[:])
}

func (oid *Oid) Copy() *Oid {
	ret := new(Oid)
	copy(ret[:], oid[:])
	return ret
}

func (oid *Oid) Equal(oid2 *Oid) bool {
	return bytes.Equal(oid[:], oid2[:])
}

func (oid *Oid) IsZero() bool {
	for _, a := range oid {
		if a != 0 {
			return false
		}
	}
	return true
}

func (oid *Oid) NCmp(oid2 *Oid, n uint) int {
	return bytes.Compare(oid[:n], oid2[:n])
}

func ShortenOids(ids []*Oid, minlen int) (int, error) {
	shorten := C.git_oid_shorten_new(C.size_t(minlen))
	if shorten == nil {
		panic("Out of memory")
	}
	defer C.git_oid_shorten_free(shorten)

	var ret C.int

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for _, id := range ids {
		buf := make([]byte, 41)
		C.git_oid_fmt((*C.char)(unsafe.Pointer(&buf[0])), id.toC())
		buf[40] = 0
		ret = C.git_oid_shorten_add(shorten, (*C.char)(unsafe.Pointer(&buf[0])))
		if ret < 0 {
			return int(ret), MakeGitError(ret)
		}
	}
	return int(ret), nil
}

type GitError struct {
	Message   string
	Class     int
	ErrorCode int
}

func (e GitError) Error() string {
	return e.Message
}

func IsNotExist(err error) bool {
	return err.(*GitError).ErrorCode == C.GIT_ENOTFOUND
}

func IsExist(err error) bool {
	return err.(*GitError).ErrorCode == C.GIT_EEXISTS
}

func MakeGitError(errorCode C.int) error {
	err := C.giterr_last()
	if err == nil {
		return &GitError{"No message", C.GITERR_INVALID, C.GIT_ERROR}
	}
	return &GitError{C.GoString(err.message), int(err.klass), int(errorCode)}
}

func MakeGitError2(err int) error {
	return MakeGitError(C.int(err))
}

func cbool(b bool) C.int {
	if b {
		return C.int(1)
	}
	return C.int(0)
}

func ucbool(b bool) C.uint {
	if b {
		return C.uint(1)
	}
	return C.uint(0)
}

func Discover(start string, across_fs bool, ceiling_dirs []string) (string, error) {
	ceildirs := C.CString(strings.Join(ceiling_dirs, string(C.GIT_PATH_LIST_SEPARATOR)))
	defer C.free(unsafe.Pointer(ceildirs))

	cstart := C.CString(start)
	defer C.free(unsafe.Pointer(cstart))

	var buf C.git_buf
	defer C.git_buf_free(&buf)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_repository_discover(&buf, cstart, cbool(across_fs), ceildirs)
	if ret < 0 {
		return "", MakeGitError(ret)
	}

	return C.GoString(buf.ptr), nil
}
