package git

/*
#include <git2.h>
#include <git2/credential.h>

git_credential_t _go_git_credential_credtype(git_credential *cred);
*/
import "C"
import (
	"fmt"
	"runtime"
	"strings"
	"unsafe"
)

// CredentialType is a bitmask of supported credential types.
//
// This represents the various types of authentication methods supported by the
// library.
type CredentialType uint

const (
	CredentialTypeUserpassPlaintext CredentialType = C.GIT_CREDENTIAL_USERPASS_PLAINTEXT
	CredentialTypeSSHKey            CredentialType = C.GIT_CREDENTIAL_SSH_KEY
	CredentialTypeSSHCustom         CredentialType = C.GIT_CREDENTIAL_SSH_CUSTOM
	CredentialTypeDefault           CredentialType = C.GIT_CREDENTIAL_DEFAULT
	CredentialTypeSSHInteractive    CredentialType = C.GIT_CREDENTIAL_SSH_INTERACTIVE
	CredentialTypeUsername          CredentialType = C.GIT_CREDENTIAL_USERNAME
	CredentialTypeSSHMemory         CredentialType = C.GIT_CREDENTIAL_SSH_MEMORY
)

func (t CredentialType) String() string {
	if t == 0 {
		return "CredentialType(0)"
	}

	var parts []string

	if (t & CredentialTypeUserpassPlaintext) != 0 {
		parts = append(parts, "UserpassPlaintext")
		t &= ^CredentialTypeUserpassPlaintext
	}
	if (t & CredentialTypeSSHKey) != 0 {
		parts = append(parts, "SSHKey")
		t &= ^CredentialTypeSSHKey
	}
	if (t & CredentialTypeSSHCustom) != 0 {
		parts = append(parts, "SSHCustom")
		t &= ^CredentialTypeSSHCustom
	}
	if (t & CredentialTypeDefault) != 0 {
		parts = append(parts, "Default")
		t &= ^CredentialTypeDefault
	}
	if (t & CredentialTypeSSHInteractive) != 0 {
		parts = append(parts, "SSHInteractive")
		t &= ^CredentialTypeSSHInteractive
	}
	if (t & CredentialTypeUsername) != 0 {
		parts = append(parts, "Username")
		t &= ^CredentialTypeUsername
	}
	if (t & CredentialTypeSSHMemory) != 0 {
		parts = append(parts, "SSHMemory")
		t &= ^CredentialTypeSSHMemory
	}

	if t != 0 {
		parts = append(parts, fmt.Sprintf("CredentialType(%#x)", t))
	}

	return strings.Join(parts, "|")
}

type Credential struct {
	ptr *C.git_credential
}

func newCredential() *Credential {
	cred := &Credential{}
	runtime.SetFinalizer(cred, (*Credential).Free)
	return cred
}

func (o *Credential) HasUsername() bool {
	if C.git_credential_has_username(o.ptr) == 1 {
		return true
	}
	return false
}

func (o *Credential) Type() CredentialType {
	return (CredentialType)(C._go_git_credential_credtype(o.ptr))
}

func (o *Credential) Free() {
	C.git_credential_free(o.ptr)
	runtime.SetFinalizer(o, nil)
	o.ptr = nil
}

func NewCredentialUsername(username string) (*Credential, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cred := newCredential()
	cusername := C.CString(username)
	ret := C.git_credential_username_new(&cred.ptr, cusername)
	if ret != 0 {
		cred.Free()
		return nil, MakeGitError(ret)
	}
	return cred, nil
}

func NewCredentialUserpassPlaintext(username string, password string) (*Credential, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cred := newCredential()
	cusername := C.CString(username)
	defer C.free(unsafe.Pointer(cusername))
	cpassword := C.CString(password)
	defer C.free(unsafe.Pointer(cpassword))
	ret := C.git_credential_userpass_plaintext_new(&cred.ptr, cusername, cpassword)
	if ret != 0 {
		cred.Free()
		return nil, MakeGitError(ret)
	}
	return cred, nil
}

// NewCredentialSSHKey creates new ssh credentials reading the public and private keys
// from the file system.
func NewCredentialSSHKey(username string, publicKeyPath string, privateKeyPath string, passphrase string) (*Credential, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cred := newCredential()
	cusername := C.CString(username)
	defer C.free(unsafe.Pointer(cusername))
	cpublickey := C.CString(publicKeyPath)
	defer C.free(unsafe.Pointer(cpublickey))
	cprivatekey := C.CString(privateKeyPath)
	defer C.free(unsafe.Pointer(cprivatekey))
	cpassphrase := C.CString(passphrase)
	defer C.free(unsafe.Pointer(cpassphrase))
	ret := C.git_credential_ssh_key_new(&cred.ptr, cusername, cpublickey, cprivatekey, cpassphrase)
	if ret != 0 {
		cred.Free()
		return nil, MakeGitError(ret)
	}
	return cred, nil
}

// NewCredentialSSHKeyFromMemory creates new ssh credentials using the publicKey and privateKey
// arguments as the values for the public and private keys.
func NewCredentialSSHKeyFromMemory(username string, publicKey string, privateKey string, passphrase string) (*Credential, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cred := newCredential()
	cusername := C.CString(username)
	defer C.free(unsafe.Pointer(cusername))
	cpublickey := C.CString(publicKey)
	defer C.free(unsafe.Pointer(cpublickey))
	cprivatekey := C.CString(privateKey)
	defer C.free(unsafe.Pointer(cprivatekey))
	cpassphrase := C.CString(passphrase)
	defer C.free(unsafe.Pointer(cpassphrase))
	ret := C.git_credential_ssh_key_memory_new(&cred.ptr, cusername, cpublickey, cprivatekey, cpassphrase)
	if ret != 0 {
		cred.Free()
		return nil, MakeGitError(ret)
	}
	return cred, nil
}

func NewCredentialSSHKeyFromAgent(username string) (*Credential, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cred := newCredential()
	cusername := C.CString(username)
	defer C.free(unsafe.Pointer(cusername))
	ret := C.git_credential_ssh_key_from_agent(&cred.ptr, cusername)
	if ret != 0 {
		cred.Free()
		return nil, MakeGitError(ret)
	}
	return cred, nil
}

func NewCredentialDefault() (*Credential, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cred := newCredential()
	ret := C.git_credential_default_new(&cred.ptr)
	if ret != 0 {
		cred.Free()
		return nil, MakeGitError(ret)
	}
	return cred, nil
}
