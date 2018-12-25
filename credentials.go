package git

/*
#include <git2.h>
#include <git2/sys/cred.h>

git_credtype_t _go_git_cred_credtype(git_cred *cred);
*/
import "C"
import (
	"errors"
	"runtime"
	"unsafe"
)

type CredType uint

const (
	CredTypeUserpassPlaintext CredType = C.GIT_CREDTYPE_USERPASS_PLAINTEXT
	CredTypeSshKey            CredType = C.GIT_CREDTYPE_SSH_KEY
	CredTypeSshCustom         CredType = C.GIT_CREDTYPE_SSH_CUSTOM
	CredTypeDefault           CredType = C.GIT_CREDTYPE_DEFAULT
	CredTypeSshInteractive    CredType = C.GIT_CREDTYPE_SSH_INTERACTIVE
	CredTypeUsername          CredType = C.GIT_CREDTYPE_USERNAME
	CredTypeSshMemory         CredType = C.GIT_CREDTYPE_SSH_MEMORY
)

type Cred struct {
	ptr *C.git_cred
}

func newCred() *Cred {
	cred := &Cred{}
	runtime.SetFinalizer(cred, (*Cred).Free)
	return cred
}

func (o *Cred) HasUsername() bool {
	if C.git_cred_has_username(o.ptr) == 1 {
		return true
	}
	return false
}

func (o *Cred) Type() CredType {
	return (CredType)(C._go_git_cred_credtype(o.ptr))
}

func (o *Cred) Free() {
	C.git_cred_free(o.ptr)
	runtime.SetFinalizer(o, nil)
	o.ptr = nil
}

// GetUserpassPlaintext returns the plaintext username/password combination stored in the Cred.
func (o *Cred) GetUserpassPlaintext() (username, password string, err error) {
	if o.Type() != CredTypeUserpassPlaintext {
		err = errors.New("credential is not userpass plaintext")
		return
	}

	plaintextCredPtr := (*C.git_cred_userpass_plaintext)(unsafe.Pointer(o.ptr))
	username = C.GoString(plaintextCredPtr.username)
	password = C.GoString(plaintextCredPtr.password)
	return
}

// GetSshKey returns the SSH-specific key information from the Cred object.
func (o *Cred) GetSshKey() (username, publickey, privatekey, passphrase string, err error) {
	if o.Type() != CredTypeSshKey && o.Type() != CredTypeSshMemory {
		err = errors.New("credential is not an ssh key")
		return
	}

	sshKeyCredPtr := (*C.git_cred_ssh_key)(unsafe.Pointer(o.ptr))
	username = C.GoString(sshKeyCredPtr.username)
	publickey = C.GoString(sshKeyCredPtr.publickey)
	privatekey = C.GoString(sshKeyCredPtr.privatekey)
	passphrase = C.GoString(sshKeyCredPtr.passphrase)
	return
}

func NewCredUserpassPlaintext(username string, password string) (*Cred, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cred := newCred()
	cusername := C.CString(username)
	defer C.free(unsafe.Pointer(cusername))
	cpassword := C.CString(password)
	defer C.free(unsafe.Pointer(cpassword))
	ret := C.git_cred_userpass_plaintext_new(&cred.ptr, cusername, cpassword)
	if ret != 0 {
		cred.Free()
		return nil, MakeGitError(ret)
	}
	return cred, nil
}

// NewCredSshKey creates new ssh credentials reading the public and private keys
// from the file system.
func NewCredSshKey(username string, publicKeyPath string, privateKeyPath string, passphrase string) (*Cred, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cred := newCred()
	cusername := C.CString(username)
	defer C.free(unsafe.Pointer(cusername))
	cpublickey := C.CString(publicKeyPath)
	defer C.free(unsafe.Pointer(cpublickey))
	cprivatekey := C.CString(privateKeyPath)
	defer C.free(unsafe.Pointer(cprivatekey))
	cpassphrase := C.CString(passphrase)
	defer C.free(unsafe.Pointer(cpassphrase))
	ret := C.git_cred_ssh_key_new(&cred.ptr, cusername, cpublickey, cprivatekey, cpassphrase)
	if ret != 0 {
		cred.Free()
		return nil, MakeGitError(ret)
	}
	return cred, nil
}

// NewCredSshKeyFromMemory creates new ssh credentials using the publicKey and privateKey
// arguments as the values for the public and private keys.
func NewCredSshKeyFromMemory(username string, publicKey string, privateKey string, passphrase string) (*Cred, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cred := newCred()
	cusername := C.CString(username)
	defer C.free(unsafe.Pointer(cusername))
	cpublickey := C.CString(publicKey)
	defer C.free(unsafe.Pointer(cpublickey))
	cprivatekey := C.CString(privateKey)
	defer C.free(unsafe.Pointer(cprivatekey))
	cpassphrase := C.CString(passphrase)
	defer C.free(unsafe.Pointer(cpassphrase))
	ret := C.git_cred_ssh_key_memory_new(&cred.ptr, cusername, cpublickey, cprivatekey, cpassphrase)
	if ret != 0 {
		cred.Free()
		return nil, MakeGitError(ret)
	}
	return cred, nil
}

func NewCredSshKeyFromAgent(username string) (*Cred, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cred := newCred()
	cusername := C.CString(username)
	defer C.free(unsafe.Pointer(cusername))
	ret := C.git_cred_ssh_key_from_agent(&cred.ptr, cusername)
	if ret != 0 {
		cred.Free()
		return nil, MakeGitError(ret)
	}
	return cred, nil
}

func NewCredDefault() (*Cred, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cred := newCred()
	ret := C.git_cred_default_new(&cred.ptr)
	if ret != 0 {
		cred.Free()
		return nil, MakeGitError(ret)
	}
	return cred, nil
}
