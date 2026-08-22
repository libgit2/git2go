package git

/*
#include <git2.h>
*/
import "C"
import (
	"runtime"
	"time"
	"unsafe"
)

type Signature struct {
	Name  string
	Email string
	When  time.Time
}

func newSignatureFromC(sig *C.git_signature) *Signature {
	if sig == nil {
		return nil
	}

	// git stores minutes, go wants seconds
	loc := time.FixedZone("", int(sig.when.offset)*60)
	return &Signature{
		C.GoString(sig.name),
		C.GoString(sig.email),
		time.Unix(int64(sig.when.time), 0).In(loc),
	}
}

// Offset returns the time zone offset of v.When in minutes, which is what git wants.
func (v *Signature) Offset() int {
	_, offset := v.When.Zone()
	return offset / 60
}

func (sig *Signature) toC() (*C.git_signature, error) {
	if sig == nil {
		return nil, nil
	}

	var out *C.git_signature

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	name := C.CString(sig.Name)
	defer C.free(unsafe.Pointer(name))

	email := C.CString(sig.Email)
	defer C.free(unsafe.Pointer(email))

	ret := C.git_signature_new(&out, name, email, C.git_time_t(sig.When.Unix()), C.int(sig.Offset()))
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return out, nil
}

func (repo *Repository) DefaultSignature() (*Signature, error) {
	var out *C.git_signature

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cErr := C.git_signature_default(&out, repo.ptr)
	runtime.KeepAlive(repo)
	if cErr < 0 {
		return nil, MakeGitError(cErr)
	}

	defer C.git_signature_free(out)

	return newSignatureFromC(out), nil
}

// DefaultSignatureFromEnv creates default author and/or committer signatures
// using environment variables and configuration.
//
// Environment variables GIT_AUTHOR_NAME, GIT_AUTHOR_EMAIL,
// GIT_COMMITTER_NAME, GIT_COMMITTER_EMAIL are honored, falling back
// to user.name and user.email configuration. For timestamps,
// GIT_AUTHOR_DATE and GIT_COMMITTER_DATE are used if set.
//
// Returns (author, committer, error). Either author or committer may be nil
// if not requested (pass false for the corresponding parameter).
func (repo *Repository) DefaultSignatureFromEnv() (author *Signature, committer *Signature, err error) {
	var authorOut *C.git_signature
	var committerOut *C.git_signature

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cErr := C.git_signature_default_from_env(&authorOut, &committerOut, repo.ptr)
	runtime.KeepAlive(repo)
	if cErr < 0 {
		return nil, nil, MakeGitError(cErr)
	}

	if authorOut != nil {
		defer C.git_signature_free(authorOut)
		author = newSignatureFromC(authorOut)
	}
	if committerOut != nil {
		defer C.git_signature_free(committerOut)
		committer = newSignatureFromC(committerOut)
	}

	return author, committer, nil
}
