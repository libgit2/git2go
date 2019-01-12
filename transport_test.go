package git

import (
	"io"
	"reflect"
	"testing"
)

type testSmartSubtransport struct {
}

func (t *testSmartSubtransport) Action(url string, action SmartServiceAction) (SmartSubtransportStream, error) {
	return &testSmartSubtransportStream{}, nil
}

func (t *testSmartSubtransport) Close() error {
	return nil
}

func (t *testSmartSubtransport) Free() {
}

type testSmartSubtransportStream struct {
}

func (s *testSmartSubtransportStream) Read(buf []byte) (int, error) {
	payload := "" +
		"001e# service=git-upload-pack\n" +
		"0000005d0000000000000000000000000000000000000000 HEAD\x00symref=HEAD:refs/heads/master agent=libgit\n" +
		"003f0000000000000000000000000000000000000000 refs/heads/master\n" +
		"0000"

	return copy(buf, []byte(payload)), io.EOF
}

func (s *testSmartSubtransportStream) Write(buf []byte) (int, error) {
	return 0, io.EOF
}

func (s *testSmartSubtransportStream) Free() {
}

func TestTransport(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	callback := func(remote *Remote, transport *Transport) (SmartSubtransport, error) {
		return &testSmartSubtransport{}, nil
	}
	registeredSmartTransport, err := NewRegisteredSmartTransport("foo", true, callback)
	checkFatal(t, err)
	defer registeredSmartTransport.Free()

	remote, err := repo.Remotes.Create("test", "foo://bar")
	checkFatal(t, err)
	defer remote.Free()

	err = remote.ConnectFetch(nil, nil, nil)
	checkFatal(t, err)

	remoteHeads, err := remote.Ls()
	checkFatal(t, err)

	expectedRemoteHeads := []RemoteHead{
		{&Oid{}, "HEAD"},
		{&Oid{}, "refs/heads/master"},
	}
	if !reflect.DeepEqual(expectedRemoteHeads, remoteHeads) {
		t.Errorf("mismatched remote heads. expected %v, got %v", expectedRemoteHeads, remoteHeads)
	}
}
