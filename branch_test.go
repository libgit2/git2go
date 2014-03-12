package git

import (
	"testing"
)

func Test_List_Branches(t *testing.T) {

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
}
