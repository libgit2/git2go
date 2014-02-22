package git

import (
	"os"
	"runtime"
	"sort"
	"testing"
	"time"
)

func TestRefModification(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	commitId, treeId := seedTestRepo(t, repo)

	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
        sig := Signature {
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
        }

        message := "this is a test"

	_, err = repo.CreateReference("refs/tags/tree", treeId, true, &sig, message)
	checkFatal(t, err)

	tag, err := repo.LookupReference("refs/tags/tree")
	checkFatal(t, err)
	checkRefType(t, tag, ReferenceOid)

	ref, err := repo.LookupReference("HEAD")
	checkFatal(t, err)
	checkRefType(t, ref, ReferenceSymbolic)

	if target := ref.Target(); target != nil {
		t.Fatalf("Expected nil *Oid, got %v", target)
	}

	ref, err = ref.Resolve()
	checkFatal(t, err)
	checkRefType(t, ref, ReferenceOid)

	if target := ref.Target(); target == nil {
		t.Fatalf("Expected valid target got nil")
	}

	if target := ref.SymbolicTarget(); target != "" {
		t.Fatalf("Expected empty string, got %v", target)
	}

	if commitId.String() != ref.Target().String() {
		t.Fatalf("Wrong ref target")
	}

	if target := ref.SymbolicTarget(); target != "" {
		t.Fatalf("Expected empty string, got %v", target)
	}

	_, err = tag.Rename("refs/tags/renamed", false, &sig, message)
	checkFatal(t, err)
	tag, err = repo.LookupReference("refs/tags/renamed")
	checkFatal(t, err)
	checkRefType(t, ref, ReferenceOid)

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

	_, err = repo.CreateReference("refs/heads/one", commitId, true, sig, message)
	checkFatal(t, err)

	_, err = repo.CreateReference("refs/heads/two", commitId, true, sig, message)
	checkFatal(t, err)

	_, err = repo.CreateReference("refs/heads/three", commitId, true, sig, message)
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
	name, err := iter.NextName()
	for err == nil {
		list = append(list, name)
		name, err = iter.NextName()
	}
	if err != ErrIterOver {
		t.Fatal("Iteration not over")
	}


	sort.Strings(list)
	compareStringList(t, expected, list)

	// test the iterator for full refs, rather than just names
	iter, err = repo.NewReferenceIterator()
	checkFatal(t, err)
	count := 0
	_, err = iter.Next()
	for err == nil {
		count++
		_, err = iter.Next()
	}
	if err != ErrIterOver {
		t.Fatal("Iteration not over")
	}

	if count != 4 {
		t.Fatalf("Wrong number of references returned %v", count)
	}


	// test the channel iteration
	list = []string{}
	iter, err = repo.NewReferenceIterator()
	for name := range iter.NameIter() {
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
	for name := range iter.NameIter() {
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

func checkRefType(t *testing.T, ref *Reference, kind ReferenceType) {
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
