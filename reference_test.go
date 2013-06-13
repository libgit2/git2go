package git

import (
	"os"
	"runtime"
	"sort"
	"testing"
)

func TestRefModification(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	commitId, treeId := seedTestRepo(t, repo)

	_, err := repo.CreateReference("refs/tags/tree", treeId, true)
	checkFatal(t, err)

	tag, err := repo.LookupReference("refs/tags/tree")
	checkFatal(t, err)
	checkRefType(t, tag, OID)

	ref, err := repo.LookupReference("HEAD")
	checkFatal(t, err)
	checkRefType(t, ref, SYMBOLIC)

	if target := ref.Target(); target != nil {
		t.Fatalf("Expected nil *Oid, got %v", target)
	}

	ref, err = ref.Resolve()
	checkFatal(t, err)
	checkRefType(t, ref, OID)

	if target := ref.Target(); target == nil {
		t.Fatalf("Expected valid target got nil")
	}

	if target := ref.SymbolicTarget(); target != "" {
		t.Fatalf("Expected empty string, got %v", target)
	}

	if commitId.String() != ref.Target().String() {
		t.Fatalf("Wrong ref target")
	}

	_, err = tag.Rename("refs/tags/renamed", false)
	checkFatal(t, err)
	tag, err = repo.LookupReference("refs/tags/renamed")
	checkFatal(t, err)
	checkRefType(t, ref, OID)

}

func TestIterator(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("README")
	checkFatal(t, err)
	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)
	commitId, err := repo.CreateCommit("HEAD", sig, sig, message, tree)
	checkFatal(t, err)

	_, err = repo.CreateReference("refs/heads/one", commitId, true)
	checkFatal(t, err)

	_, err = repo.CreateReference("refs/heads/two", commitId, true)
	checkFatal(t, err)

	_, err = repo.CreateReference("refs/heads/three", commitId, true)
	checkFatal(t, err)

	iter, err := repo.NewReferenceIterator()
	checkFatal(t, err)

	var list []string
	expected := []string{
		"refs/heads/master",
		"refs/heads/one",
		"refs/heads/three",
		"refs/heads/two",
	}

	// test some manual iteration
	name, err := iter.Next()
	for err == nil {
		list = append(list, name)
		name, err = iter.Next()
	}
	if err != ErrIterOver {
		t.Fatal("Iteration not over")
	}


	sort.Strings(list)
	compareStringList(t, expected, list)

	// test the channel iteration
	list = []string{}
	iter, err = repo.NewReferenceIterator()
	for name := range iter.Iter() {
		list = append(list, name)
	}

	sort.Strings(list)
	compareStringList(t, expected, list)

	iter, err = repo.NewReferenceIteratorGlob("refs/heads/t*")
	expected = []string{
		"refs/heads/three",
		"refs/heads/two",
	}

	list = []string{}
	for name := range iter.Iter() {
		list = append(list, name)
	}

	compareStringList(t, expected, list)
}

func compareStringList(t *testing.T, expected, actual []string) {
	for i, v := range expected {
		if actual[i] != v {
			t.Fatalf("Bad list")
		}
	}
}

func checkRefType(t *testing.T, ref *Reference, kind int) {
	if ref.Type() == kind {
		return
	}

	// The failure happens at wherever we were called, not here
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatal()
	}

	t.Fatalf("Wrong ref type at %v:%v; have %v, expected %v", file, line, ref.Type(), kind)
}
