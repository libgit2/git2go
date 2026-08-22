package git

import (
	"testing"
)

func TestBranchIterator(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	i, err := repo.NewBranchIterator(BranchLocal)
	checkFatal(t, err)

	b, bt, err := i.Next()
	checkFatal(t, err)
	branchName := defaultBranchName(t, repo)
	if name, _ := b.Name(); name != branchName {
		t.Fatalf("expected %s, got %s", branchName, name)
	} else if bt != BranchLocal {
		t.Fatalf("expected BranchLocal, not %v", t)
	}
	b, bt, err = i.Next()
	if !IsErrorCode(err, ErrorCodeIterOver) {
		t.Fatal("expected iterover")
	}
}

func TestBranchIteratorEach(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	i, err := repo.NewBranchIterator(BranchLocal)
	checkFatal(t, err)

	var names []string
	f := func(b *Branch, t BranchType) error {
		name, err := b.Name()
		if err != nil {
			return err
		}

		names = append(names, name)
		return nil
	}

	err = i.ForEach(f)
	if err != nil && !IsErrorCode(err, ErrorCodeIterOver) {
		t.Fatal(err)
	}

	if len(names) != 1 {
		t.Fatalf("expect 1 branch, but it was %d\n", len(names))
	}

	branchName := defaultBranchName(t, repo)
	if names[0] != branchName {
		t.Fatalf("expect branch %s, but it was %s\n", branchName, names[0])
	}
}
