package git

import (
	"io/ioutil"
	"path"
	"testing"
)

func TestStatusFile(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	state := repo.State()
	if state != RepositoryStateNone {
		t.Fatal("Incorrect repository state: ", state)
	}

	err := ioutil.WriteFile(path.Join(path.Dir(repo.Workdir()), "hello.txt"), []byte("Hello, World"), 0644)
	checkFatal(t, err)

	status, err := repo.StatusFile("hello.txt")
	checkFatal(t, err)

	if status != StatusWtNew {
		t.Fatal("Incorrect status flags: ", status)
	}
}

func TestStatusList(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	// This commits the test repo README, so it doesn't show up in the status list and there's a head to compare to
	seedTestRepo(t, repo)

	err := ioutil.WriteFile(path.Join(path.Dir(repo.Workdir()), "hello.txt"), []byte("Hello, World"), 0644)
	checkFatal(t, err)

	opts := &StatusOptions{}
	opts.Show = StatusShowIndexAndWorkdir
	opts.Flags = StatusOptIncludeUntracked | StatusOptRenamesHeadToIndex | StatusOptSortCaseSensitively

	statusList, err := repo.StatusList(opts)
	checkFatal(t, err)

	entryCount, err := statusList.EntryCount()
	checkFatal(t, err)

	if entryCount != 1 {
		t.Fatal("Incorrect number of status entries: ", entryCount)
	}

	entry, err := statusList.ByIndex(0)
	checkFatal(t, err)
	if entry.Status != StatusWtNew {
		t.Fatal("Incorrect status flags: ", entry.Status)
	}
	if entry.IndexToWorkdir.NewFile.Path != "hello.txt" {
		t.Fatal("Incorrect entry path: ", entry.IndexToWorkdir.NewFile.Path)
	}
}

func TestStatusNothing(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	opts := &StatusOptions{
		Show:  StatusShowIndexAndWorkdir,
		Flags: StatusOptIncludeUntracked | StatusOptRenamesHeadToIndex | StatusOptSortCaseSensitively,
	}

	statusList, err := repo.StatusList(opts)
	checkFatal(t, err)

	entryCount, err := statusList.EntryCount()
	checkFatal(t, err)

	if entryCount != 0 {
		t.Fatal("expected no statuses in empty repo")
	}

	_, err = statusList.ByIndex(0)
	if err == nil {
		t.Error("expected error getting status by index")
	}
}
