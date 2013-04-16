package git

import (
	"os"
	"runtime"
	"testing"
	"time"
)

func seedTestRepo(t *testing.T, repo *Repository) (*Oid, *Oid) {
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

	return commitId, treeId
}

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
