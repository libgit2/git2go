package git

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/shlex"
	"golang.org/x/crypto/ssh"
)

func TestListRemotes(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	remote, err := repo.Remotes.Create("test", "git://foo/bar")
	checkFatal(t, err)
	defer remote.Free()

	expected := []string{
		"test",
	}

	actual, err := repo.Remotes.List()
	checkFatal(t, err)

	compareStringList(t, expected, actual)
}

func assertHostname(cert *Certificate, valid bool, hostname string, t *testing.T) ErrorCode {
	if hostname != "github.com" {
		t.Fatal("Hostname does not match")
		return ErrorCodeUser
	}

	return ErrorCodeOK
}

func TestCertificateCheck(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	remote, err := repo.Remotes.Create("origin", "https://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)
	defer remote.Free()

	options := FetchOptions{
		RemoteCallbacks: RemoteCallbacks{
			CertificateCheckCallback: func(cert *Certificate, valid bool, hostname string) ErrorCode {
				return assertHostname(cert, valid, hostname, t)
			},
		},
	}

	err = remote.Fetch([]string{}, &options, "")
	checkFatal(t, err)
}

func TestRemoteConnect(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	remote, err := repo.Remotes.Create("origin", "https://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)
	defer remote.Free()

	err = remote.ConnectFetch(nil, nil, nil)
	checkFatal(t, err)
}

func TestRemoteConnectOption(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	config, err := repo.Config()
	checkFatal(t, err)
	err = config.SetString("url.git@github.com:.insteadof", "https://github.com/")
	checkFatal(t, err)

	option, err := DefaultRemoteCreateOptions()
	checkFatal(t, err)
	option.Name = "origin"
	option.Flags = RemoteCreateSkipInsteadof

	remote, err := repo.Remotes.CreateWithOptions("https://github.com/libgit2/TestGitRepository", option)
	checkFatal(t, err)
	defer remote.Free()

	err = remote.ConnectFetch(nil, nil, nil)
	checkFatal(t, err)
}

func TestRemoteLs(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	remote, err := repo.Remotes.Create("origin", "https://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)
	defer remote.Free()

	err = remote.ConnectFetch(nil, nil, nil)
	checkFatal(t, err)

	heads, err := remote.Ls()
	checkFatal(t, err)

	if len(heads) == 0 {
		t.Error("Expected remote heads")
	}
}

func TestRemoteLsFiltering(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	remote, err := repo.Remotes.Create("origin", "https://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)
	defer remote.Free()

	err = remote.ConnectFetch(nil, nil, nil)
	checkFatal(t, err)

	heads, err := remote.Ls("master")
	checkFatal(t, err)

	if len(heads) != 1 {
		t.Fatalf("Expected one head for master but I got %d", len(heads))
	}

	if heads[0].Id == nil {
		t.Fatalf("Expected head to have an Id, but it's nil")
	}

	if heads[0].Name == "" {
		t.Fatalf("Expected head to have a name, but it's empty")
	}
}

func TestRemotePruneRefs(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	config, err := repo.Config()
	checkFatal(t, err)
	defer config.Free()

	err = config.SetBool("remote.origin.prune", true)
	checkFatal(t, err)

	remote, err := repo.Remotes.Create("origin", "https://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)
	defer remote.Free()

	remote, err = repo.Remotes.Lookup("origin")
	checkFatal(t, err)
	defer remote.Free()

	if !remote.PruneRefs() {
		t.Fatal("Expected remote to be configured to prune references")
	}
}

func TestRemotePrune(t *testing.T) {
	t.Parallel()
	remoteRepo := createTestRepo(t)
	defer cleanupTestRepo(t, remoteRepo)

	head, _ := seedTestRepo(t, remoteRepo)
	commit, err := remoteRepo.LookupCommit(head)
	checkFatal(t, err)
	defer commit.Free()

	remoteRef, err := remoteRepo.CreateBranch("test-prune", commit, true)
	checkFatal(t, err)
	defer remoteRef.Free()

	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	config, err := repo.Config()
	checkFatal(t, err)
	defer config.Free()

	remoteUrl := fmt.Sprintf("file://%s", remoteRepo.Workdir())
	remote, err := repo.Remotes.Create("origin", remoteUrl)
	checkFatal(t, err)
	defer remote.Free()

	err = remote.Fetch([]string{"test-prune"}, nil, "")
	checkFatal(t, err)

	ref, err := repo.References.Create("refs/remotes/origin/test-prune", head, true, "remote reference")
	checkFatal(t, err)
	defer ref.Free()

	err = remoteRef.Delete()
	checkFatal(t, err)

	err = config.SetBool("remote.origin.prune", true)
	checkFatal(t, err)

	rr, err := repo.Remotes.Lookup("origin")
	checkFatal(t, err)
	defer rr.Free()

	err = rr.ConnectFetch(nil, nil, nil)
	checkFatal(t, err)

	err = rr.Prune(nil)
	checkFatal(t, err)

	ref, err = repo.References.Lookup("refs/remotes/origin/test-prune")
	if err == nil {
		ref.Free()
		t.Fatal("Expected error getting a pruned reference")
	}
}

func newChannelPipe(t *testing.T, w io.Writer, wg *sync.WaitGroup) (*os.File, error) {
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	wg.Add(1)
	go func() {
		_, err := io.Copy(w, pr)
		if err != nil && err != io.EOF {
			t.Logf("Failed to copy: %v", err)
		}
		wg.Done()
	}()

	return pw, nil
}

func startSSHServer(t *testing.T, hostKey ssh.Signer, authorizedKeys []ssh.PublicKey) net.Listener {
	t.Helper()

	marshaledAuthorizedKeys := make([][]byte, len(authorizedKeys))
	for i, authorizedKey := range authorizedKeys {
		marshaledAuthorizedKeys[i] = authorizedKey.Marshal()
	}

	config := &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			marshaledPubKey := pubKey.Marshal()
			for _, marshaledAuthorizedKey := range marshaledAuthorizedKeys {
				if bytes.Equal(marshaledPubKey, marshaledAuthorizedKey) {
					return &ssh.Permissions{
						// Record the public key used for authentication.
						Extensions: map[string]string{
							"pubkey-fp": ssh.FingerprintSHA256(pubKey),
						},
					}, nil
				}
			}
			t.Logf("unknown public key for %q:\n\t%+v\n\t%+v\n", c.User(), pubKey.Marshal(), authorizedKeys)
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}
	config.AddHostKey(hostKey)

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen for connection: %v", err)
	}

	go func() {
		nConn, err := listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			t.Logf("Failed to accept incoming connection: %v", err)
			return
		}
		defer nConn.Close()

		conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			t.Logf("failed to handshake: %+v, %+v", conn, err)
			return
		}

		// The incoming Request channel must be serviced.
		go func() {
			for newRequest := range reqs {
				t.Logf("new request %v", newRequest)
			}
		}()

		// Service only the first channel request
		newChannel := <-chans
		defer func() {
			for newChannel := range chans {
				t.Logf("new channel %v", newChannel)
				newChannel.Reject(ssh.UnknownChannelType, "server closing")
			}
		}()

		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			return
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			t.Logf("Could not accept channel: %v", err)
			return
		}
		defer channel.Close()

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "exec" request.
		req := <-requests
		if req.Type != "exec" {
			req.Reply(false, nil)
			return
		}
		// RFC 4254 Section 6.5.
		var payload struct {
			Command string
		}
		if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
			t.Logf("invalid payload on channel %v: %v", channel, err)
			req.Reply(false, nil)
			return
		}
		args, err := shlex.Split(payload.Command)
		if err != nil {
			t.Logf("invalid command on channel %v: %v", channel, err)
			req.Reply(false, nil)
			return
		}
		if len(args) < 2 || (args[0] != "git-upload-pack" && args[0] != "git-receive-pack") {
			t.Logf("invalid command (%v) on channel %v: %v", args, channel, err)
			req.Reply(false, nil)
			return
		}
		req.Reply(true, nil)

		go func(in <-chan *ssh.Request) {
			for req := range in {
				t.Logf("draining request %v", req)
			}
		}(requests)

		// The first parameter is the (absolute) path of the repository.
		args[1] = "./testdata" + args[1]

		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = channel
		var wg sync.WaitGroup
		stdoutPipe, err := newChannelPipe(t, channel, &wg)
		if err != nil {
			t.Logf("Failed to create stdout pipe: %v", err)
			return
		}
		cmd.Stdout = stdoutPipe
		stderrPipe, err := newChannelPipe(t, channel.Stderr(), &wg)
		if err != nil {
			t.Logf("Failed to create stderr pipe: %v", err)
			return
		}
		cmd.Stderr = stderrPipe

		go func() {
			wg.Wait()
			channel.CloseWrite()
		}()

		err = cmd.Start()
		if err != nil {
			t.Logf("Failed to start %v: %v", args, err)
			return
		}

		// Once the process has started, we need to close the write end of the
		// pipes from this process so that we can know when the child has done
		// writing to it.
		stdoutPipe.Close()
		stderrPipe.Close()

		timer := time.AfterFunc(5*time.Second, func() {
			t.Log("process timed out, terminating")
			cmd.Process.Kill()
		})
		defer timer.Stop()

		err = cmd.Wait()
		if err != nil {
			t.Logf("Failed to run %v: %v", args, err)
			return
		}
	}()
	return listener
}

func TestRemoteSSH(t *testing.T) {
	t.Parallel()
	pubKeyUsername := "testuser"

	hostPrivKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatalf("Failed to generate the host RSA private key: %v", err)
	}
	hostSigner, err := ssh.NewSignerFromKey(hostPrivKey)
	if err != nil {
		t.Fatalf("Failed to generate SSH hostSigner: %v", err)
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatalf("Failed to generate the user RSA private key: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(privKey)
	if err != nil {
		t.Fatalf("Failed to generate SSH signer: %v", err)
	}
	// This is in the format "xx:xx:xx:...", so we remove the colons so that it
	// matches the fmt.Sprintf() below.
	// Note that not all libssh2 implementations support the SHA256 fingerprint,
	// so we use MD5 here for testing.
	publicKeyFingerprint := strings.Replace(ssh.FingerprintLegacyMD5(hostSigner.PublicKey()), ":", "", -1)

	listener := startSSHServer(t, hostSigner, []ssh.PublicKey{signer.PublicKey()})
	defer listener.Close()

	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	certificateCheckCallbackCalled := false
	fetchOpts := FetchOptions{
		RemoteCallbacks: RemoteCallbacks{
			CertificateCheckCallback: func(cert *Certificate, valid bool, hostname string) ErrorCode {
				hostkeyFingerprint := fmt.Sprintf("%x", cert.Hostkey.HashMD5[:])
				if hostkeyFingerprint != publicKeyFingerprint {
					t.Logf("server hostkey %q, want %q", hostkeyFingerprint, publicKeyFingerprint)
					return ErrorCodeAuth
				}
				certificateCheckCallbackCalled = true
				return ErrorCodeOK
			},
			CredentialsCallback: func(url, username string, allowedTypes CredentialType) (*Credential, error) {
				if allowedTypes&(CredentialTypeSSHKey|CredentialTypeSSHCustom|CredentialTypeSSHMemory) != 0 {
					return NewCredentialSSHKeyFromSigner(pubKeyUsername, signer)
				}
				if (allowedTypes & CredentialTypeUsername) != 0 {
					return NewCredentialUsername(pubKeyUsername)
				}
				return nil, fmt.Errorf("unknown credential type %+v", allowedTypes)
			},
		},
	}

	remote, err := repo.Remotes.Create(
		"origin",
		fmt.Sprintf("ssh://%s/TestGitRepository", listener.Addr().String()),
	)
	checkFatal(t, err)
	defer remote.Free()

	err = remote.Fetch(nil, &fetchOpts, "")
	checkFatal(t, err)
	if !certificateCheckCallbackCalled {
		t.Fatalf("CertificateCheckCallback was not called")
	}

	heads, err := remote.Ls()
	checkFatal(t, err)

	if len(heads) == 0 {
		t.Error("Expected remote heads")
	}
}
