package git

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"runtime"
)

func registerManagedSsh() error {
	registeredSmartTransport, err := NewRegisteredSmartTransport("ssh", false, sshSmartSubtransportFactory)
	if err != nil {
		registeredSmartTransport.Free()
	}
	return err
}

func sshSmartSubtransportFactory(remote *Remote, transport *Transport) (SmartSubtransport, error) {
	return &sshSmartSubtransport{
		transport: transport,
	}, nil
}

type sshSmartSubtransport struct {
	transport *Transport

	action  SmartServiceAction
	client  *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser
	stdout  io.Reader
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
		if t.client != nil {
			if t.action == SmartServiceActionUploadpackLs || t.action == SmartServiceActionUploadpack {
				return &sshSmartSubtransportStream{
					owner: t,
				}, nil
			}
			t.Close()
		}
		cmd = fmt.Sprintf("git-upload-pack %q", u.Path)

	case SmartServiceActionReceivepackLs, SmartServiceActionReceivepack:
		if t.client != nil {
			if t.action == SmartServiceActionReceivepackLs || t.action == SmartServiceActionReceivepack {
				return &sshSmartSubtransportStream{
					owner: t,
				}, nil
			}
			t.Close()
		}
		cmd = fmt.Sprintf("git-receive-pack %q", u.Path)

	default:
		return nil, err
	}

	cred, err := t.transport.SmartCredentials("", CredTypeSshKey|CredTypeSshMemory)
	if err != nil {
		return nil, err
	}
	defer cred.Free()

	sshConfig, err := getSshConfigFromCred(cred)
	if err != nil {
		return nil, err
	}
	sshConfig.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		marshaledKey := key.Marshal()
		cert := &Certificate{
			Kind: CertificateHostkey,
			Hostkey: HostkeyCertificate{
				Kind:     HostkeySHA1 | HostkeyMD5,
				HashMD5:  md5.Sum(marshaledKey),
				HashSHA1: sha1.Sum(marshaledKey),
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

	t.action = action
	return &sshSmartSubtransportStream{
		owner: t,
	}, nil
}

func (t *sshSmartSubtransport) Close() error {
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

func getSshConfigFromCred(cred *Cred) (*ssh.ClientConfig, error) {
	username, _, privatekey, passphrase, err := cred.GetSshKey()
	if err != nil {
		return nil, err
	}

	var pemBytes []byte
	if cred.Type() == CredTypeSshMemory {
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
