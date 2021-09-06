package git

/*
#include <git2.h>

#include <git2/sys/credential.h>
*/
import "C"
import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"runtime"
	"unsafe"

	"golang.org/x/crypto/ssh"
)

// RegisterManagedSSHTransport registers a Go-native implementation of an SSH
// transport that doesn't rely on any system libraries (e.g. libssh2).
//
// If Shutdown or ReInit are called, make sure that the smart transports are
// freed before it.
func RegisterManagedSSHTransport(protocol string) (*RegisteredSmartTransport, error) {
	return NewRegisteredSmartTransport(protocol, false, sshSmartSubtransportFactory)
}

func registerManagedSSH() error {
	globalRegisteredSmartTransports.Lock()
	defer globalRegisteredSmartTransports.Unlock()

	for _, protocol := range []string{"ssh", "ssh+git", "git+ssh"} {
		if _, ok := globalRegisteredSmartTransports.transports[protocol]; ok {
			continue
		}
		managed, err := newRegisteredSmartTransport(protocol, false, sshSmartSubtransportFactory, true)
		if err != nil {
			return fmt.Errorf("failed to register transport for %q: %v", protocol, err)
		}
		globalRegisteredSmartTransports.transports[protocol] = managed
	}
	return nil
}

func sshSmartSubtransportFactory(remote *Remote, transport *Transport) (SmartSubtransport, error) {
	return &sshSmartSubtransport{
		transport: transport,
	}, nil
}

type sshSmartSubtransport struct {
	transport *Transport

	lastAction    SmartServiceAction
	client        *ssh.Client
	session       *ssh.Session
	stdin         io.WriteCloser
	stdout        io.Reader
	currentStream *sshSmartSubtransportStream
}

func (t *sshSmartSubtransport) Action(urlString string, action SmartServiceAction) (SmartSubtransportStream, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	u, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	var cmd string
	switch action {
	case SmartServiceActionUploadpackLs, SmartServiceActionUploadpack:
		if t.currentStream != nil {
			if t.lastAction == SmartServiceActionUploadpackLs {
				return t.currentStream, nil
			}
			t.Close()
		}
		cmd = fmt.Sprintf("git-upload-pack %q", u.Path)

	case SmartServiceActionReceivepackLs, SmartServiceActionReceivepack:
		if t.currentStream != nil {
			if t.lastAction == SmartServiceActionReceivepackLs {
				return t.currentStream, nil
			}
			t.Close()
		}
		cmd = fmt.Sprintf("git-receive-pack %q", u.Path)

	default:
		return nil, fmt.Errorf("unexpected action: %v", action)
	}

	cred, err := t.transport.SmartCredentials("", CredentialTypeSSHKey|CredentialTypeSSHMemory)
	if err != nil {
		return nil, err
	}
	defer cred.Free()

	sshConfig, err := getSSHConfigFromCredential(cred)
	if err != nil {
		return nil, err
	}
	sshConfig.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		marshaledKey := key.Marshal()
		cert := &Certificate{
			Kind: CertificateHostkey,
			Hostkey: HostkeyCertificate{
				Kind:         HostkeySHA1 | HostkeyMD5 | HostkeySHA256 | HostkeyRaw,
				HashMD5:      md5.Sum(marshaledKey),
				HashSHA1:     sha1.Sum(marshaledKey),
				HashSHA256:   sha256.Sum256(marshaledKey),
				Hostkey:      marshaledKey,
				SSHPublicKey: key,
			},
		}

		return t.transport.SmartCertificateCheck(cert, true, hostname)
	}

	var addr string
	if u.Port() != "" {
		addr = fmt.Sprintf("%s:%s", u.Hostname(), u.Port())
	} else {
		addr = fmt.Sprintf("%s:22", u.Hostname())
	}

	t.client, err = ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, err
	}

	t.session, err = t.client.NewSession()
	if err != nil {
		return nil, err
	}

	t.stdin, err = t.session.StdinPipe()
	if err != nil {
		return nil, err
	}

	t.stdout, err = t.session.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := t.session.Start(cmd); err != nil {
		return nil, err
	}

	t.lastAction = action
	t.currentStream = &sshSmartSubtransportStream{
		owner: t,
	}

	return t.currentStream, nil
}

func (t *sshSmartSubtransport) Close() error {
	t.currentStream = nil
	if t.client != nil {
		t.stdin.Close()
		t.session.Wait()
		t.session.Close()
		t.client = nil
	}
	return nil
}

func (t *sshSmartSubtransport) Free() {
}

type sshSmartSubtransportStream struct {
	owner *sshSmartSubtransport
}

func (stream *sshSmartSubtransportStream) Read(buf []byte) (int, error) {
	return stream.owner.stdout.Read(buf)
}

func (stream *sshSmartSubtransportStream) Write(buf []byte) (int, error) {
	return stream.owner.stdin.Write(buf)
}

func (stream *sshSmartSubtransportStream) Free() {
}

func getSSHConfigFromCredential(cred *Credential) (*ssh.ClientConfig, error) {
	switch cred.Type() {
	case CredentialTypeSSHCustom:
		credSSHCustom := (*C.git_credential_ssh_custom)(unsafe.Pointer(cred.ptr))
		data, ok := pointerHandles.Get(credSSHCustom.payload).(*credentialSSHCustomData)
		if !ok {
			return nil, errors.New("unsupported custom SSH credentials")
		}
		return &ssh.ClientConfig{
			User: C.GoString(credSSHCustom.username),
			Auth: []ssh.AuthMethod{ssh.PublicKeys(data.signer)},
		}, nil
	}

	username, _, privatekey, passphrase, err := cred.GetSSHKey()
	if err != nil {
		return nil, err
	}

	var pemBytes []byte
	if cred.Type() == CredentialTypeSSHMemory {
		pemBytes = []byte(privatekey)
	} else {
		pemBytes, err = ioutil.ReadFile(privatekey)
		if err != nil {
			return nil, err
		}
	}

	var key ssh.Signer
	if passphrase != "" {
		key, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(passphrase))
		if err != nil {
			return nil, err
		}
	} else {
		key, err = ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			return nil, err
		}
	}

	return &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(key)},
	}, nil
}
