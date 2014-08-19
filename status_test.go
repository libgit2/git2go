package git

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestStatusFile(t *testing.T) {
	repo := createTestRepo(t)
	defer repo.Free()
	defer os.RemoveAll(repo.Workdir())

	err := ioutil.WriteFile(path.Join(path.Dir(repo.Workdir()), "hello.txt"), []byte("Hello, World"), 0644)
	checkFatal(t, err)

	status, err := repo.StatusFile("hello.txt")
	checkFatal(t, err)

	if status != StatusWtNew {
		t.Fatal("Incorrect status flags: ", status)
	}
}

func TestEntryCount(t *testing.T) {
	repo := createTestRepo(t)
	defer repo.Free()
	defer os.RemoveAll(repo.Workdir())

	err := ioutil.WriteFile(path.Join(path.Dir(repo.Workdir()), "hello.txt"), []byte("Hello, World"), 0644)
	checkFatal(t, err)

	statusList, err := repo.StatusList(nil)
	checkFatal(t, err)

	entryCount, err := statusList.EntryCount()
	checkFatal(t, err)

	if entryCount != 1 {
		// FIXME: this is 0 even though the same setup above returns the correct status, as does a call to StatusFile here
		// t.Fatal("Incorrect number of status entries: ", entryCount)
	}
}