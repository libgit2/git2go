package git

import (
	"os"
	"runtime"
	"testing"
	"io/ioutil"
)

func TestCreateRepoAndStage(t *testing.T) {
	// figure out where we can create the test repo
	path, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)
	repo, err := InitRepository(path, false)
	checkFatal(t, err)

	tmpfile := "README"
	err = ioutil.WriteFile(path + "/" + tmpfile, []byte("foo\n"), 0644)
	checkFatal(t, err)
	defer os.RemoveAll(path)

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath(tmpfile)
	checkFatal(t, err)
	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	if treeId.String() != "b7119b11e8ef7a1a5a34d3ac87f5b075228ac81e" {
		t.Fatalf("%v", treeId.String())
	}
}

func checkFatal(t *testing.T, err error) {
	if err == nil {
		return
	}

	// The failure happens at wherever we were called, not here
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatal()
	}

	t.Fatalf("Fail at %v:%v", file, line)
}
