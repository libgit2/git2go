package git

import (
	"testing"
)

func TestRemoteLs(t *testing.T) {
	repo := createTestRepo(t)
	remote, err := repo.CreateRemote("origin", "git://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)

	err = remote.Connect(RemoteDirectionFetch)
	checkFatal(t, err)
	
	if remote.IsConnected() != true {
		t.Fatal("Connected but not connected")
	}

	expected := []string{
		"HEAD",
		"refs/heads/first-merge",
		"refs/heads/master",
		"refs/heads/no-parent",
		"refs/tags/annotated_tag",
		"refs/tags/annotated_tag^{}",
		"refs/tags/blob",
		"refs/tags/commit_tree",
		"refs/tags/nearly-dangling",
	}

	refs, err := remote.Ls()
	for i, s := range expected {
		if refs[i].Name != s {
			t.Fatal("remote refs not as expected")
		}
	}
}