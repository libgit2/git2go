package git

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestEntryCount(t *testing.T) {
	repo := createTestRepo(t)
	defer repo.Free()
	defer os.RemoveAll(repo.Workdir())

	err := ioutil.WriteFile(path.Join(path.Dir(repo.Path()), "hello.txt"), []byte("Hello, World"), 0644)
	checkFatal(t, err)

	statusList, err := repo.StatusList()
	checkFatal(t, err)

	entryCount, err := statusList.EntryCount()
	checkFatal(t, err)

	if entryCount != 1 {
		t.Fatal("Incorrect number of status entries: ", entryCount)
	}
}
