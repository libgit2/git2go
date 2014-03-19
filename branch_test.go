package git

import (
	"testing"
)

func TestBranchIterator(t *testing.T) {

	repo := createTestRepo(t)
	seedTestRepo(t, repo)

	i, err := repo.NewBranchIterator(BranchLocal)
	checkFatal(t, err)

	ref, err := i.Next()
	checkFatal(t, err)
	if ref.Name() != "refs/heads/master" {
		t.Fatalf("expected refs/heads/master, not %v", ref.Name())
	}
	ref, err = i.Next()
	if ref != nil {
		t.Fatal("expected nil")
	}

	if err != ErrIterOver {
		t.Fatal("expected iterover")
	}

	// test channel iterator

	i, err = repo.NewBranchIterator(BranchLocal)
	checkFatal(t, err)

	list := make([]string, 0)
	for ref := range NameIteratorChannel(i) {
		list = append(list, ref)
	}

	if len(list) != 1 {
		t.Fatal("expected single match")
	}

	if list[0] != "refs/heads/master" {
		t.Fatal("expected refs/heads/master")
	}

}
