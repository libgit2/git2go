package git

/*
#include <git2.h>
#include <git2/errors.h>
*/
import "C"
import (
	"unsafe"
	"runtime"
)

type CredType uint

const (
	CredTypeUserpassPlaintext CredType = C.GIT_CREDTYPE_USERPASS_PLAINTEXT
	CredTypeSshKey                     = C.GIT_CREDTYPE_SSH_KEY
	CredTypeSshCustom                  = C.GIT_CREDTYPE_SSH_CUSTOM
	CredTypeDefault                    = C.GIT_CREDTYPE_DEFAULT
)

type Cred struct {
	ptr *C.git_cred
}

func (o *Cred) HasUsername() bool {
	return C.git_cred_has_username(o.ptr) != 0
}

func (o *Cred) Type() CredType {
	return CredType(o.ptr.credtype)
}

func credFromC(ptr *C.git_cred) *Cred {
	return &Cred{ptr}
}

func NewCredUserpassPlaintext(username string, password string) (*Cred, error) {
	cred := Cred{}
	cusername := C.CString(username)
	defer C.free(unsafe.Pointer(cusername))
	cpassword := C.CString(password)
	defer C.free(unsafe.Pointer(cpassword))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_cred_userpass_plaintext_new(&cred.ptr, cusername, cpassword)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}


	return &cred, nil
}

func NewCredSshKey(username, publickey, privatekey, passphrase string) (*Cred, error) {
	cred := Cred{}
	cusername := C.CString(username)
	defer C.free(unsafe.Pointer(cusername))
	cpublickey := C.CString(publickey)
	defer C.free(unsafe.Pointer(cpublickey))
	cprivatekey := C.CString(privatekey)
	defer C.free(unsafe.Pointer(cprivatekey))
	cpassphrase := C.CString(passphrase)
	defer C.free(unsafe.Pointer(cpassphrase))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_cred_ssh_key_new(&cred.ptr, cusername, cpublickey, cprivatekey, cpassphrase)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return &cred, nil
}

func NewCredSshKeyFromAgent(username string) (*Cred, error) {
	cred := Cred{}
	cusername := C.CString(username)
	defer C.free(unsafe.Pointer(cusername))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_cred_ssh_key_from_agent(&cred.ptr, cusername)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return &cred, nil
}

func NewCredDefault() (*Cred, error) {
	cred := Cred{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_cred_default_new(&cred.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return &cred, nil
}
