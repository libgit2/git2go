package git

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func assertSamePath(t *testing.T, expected string, actual string) {
	var err error
	expected, err = filepath.EvalSymlinks(expected)
	checkFatal(t, err)
	actual, err = filepath.EvalSymlinks(actual)
	checkFatal(t, err)

	if expected != actual {
		t.Fatalf("wrong path (expected %s, got %s)", expected, actual)
	}
}

func TestAddWorkspace(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)

	worktreeTemporaryPath, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)
	defer func() { checkFatal(t, os.RemoveAll(worktreeTemporaryPath)) }()
	worktreeName := "testWorktree"
	worktreePath := filepath.Join(worktreeTemporaryPath, "worktree")

	worktree, err := repo.Worktrees.Add(worktreeName, worktreePath, &AddWorktreeOptions{
		Lock: true, CheckoutOptions: CheckoutOptions{Strategy: CheckoutForce},
	})
	checkFatal(t, err)

	if name := worktree.Name(); name != worktreeName {
		t.Fatalf("wrong worktree name: %s != %s", worktreeName, name)
	}
	locked, _, err := worktree.IsLocked()
	checkFatal(t, err)
	if locked != true {
		t.Fatal("worktree isn't locked")
	}
	assertSamePath(t, worktreePath, worktree.Path())
	checkFatal(t, worktree.Validate())
}

func TestAddWorkspaceWithoutOptions(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)

	worktreeTemporaryPath, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)
	defer func() { checkFatal(t, os.RemoveAll(worktreeTemporaryPath)) }()
	worktreeName := "testWorktree"
	worktreePath := filepath.Join(worktreeTemporaryPath, "worktree")

	worktree, err := repo.Worktrees.Add(worktreeName, worktreePath, nil)
	checkFatal(t, err)

	if name := worktree.Name(); name != worktreeName {
		t.Fatalf("wrong worktree name: %s != %s", worktreeName, name)
	}
	locked, _, err := worktree.IsLocked()
	checkFatal(t, err)
	if locked != false {
		t.Fatal("worktree is locked")
	}
	assertSamePath(t, worktreePath, worktree.Path())
	checkFatal(t, worktree.Validate())
}

func TestLookupWorkspace(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)

	worktreeTemporaryPath, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)
	defer func() { checkFatal(t, os.RemoveAll(worktreeTemporaryPath)) }()
	worktreeName := "testWorktree"

	worktree, err := repo.Worktrees.Add(worktreeName, filepath.Join(worktreeTemporaryPath, "worktree"), nil)
	checkFatal(t, err)
	retrievedWorktree, err := repo.Worktrees.Lookup(worktreeName)
	checkFatal(t, err)

	assertSamePath(t, worktree.Path(), retrievedWorktree.Path())
}

func TestListWorkspaces(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)

	worktreeTemporaryPath, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)
	defer func() { checkFatal(t, os.RemoveAll(worktreeTemporaryPath)) }()

	worktreeNames := []string{"worktree1", "worktree2", "worktree3"}
	for _, name := range worktreeNames {
		_, err = repo.Worktrees.Add(name, filepath.Join(worktreeTemporaryPath, name), nil)
		checkFatal(t, err)
	}
	listedWorktree, err := repo.Worktrees.List()
	checkFatal(t, err)

	if len(worktreeNames) != len(listedWorktree) {
		t.Fatalf("len(worktreeNames) != len(listedWorktree) as %d != %d", len(worktreeNames), len(listedWorktree))
	}
	for _, name := range worktreeNames {
		found := false
		for _, nameToMatch := range listedWorktree {
			if name == nameToMatch {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("worktree %s is missing", name)
		}
	}
}

func TestOpenWorkspaceFromRepository(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)

	worktreeTemporaryPath, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)
	defer func() { checkFatal(t, os.RemoveAll(worktreeTemporaryPath)) }()

	worktree, err := repo.Worktrees.Add("testWorktree", filepath.Join(worktreeTemporaryPath, "worktree"), nil)
	checkFatal(t, err)
	worktreeRepo, err := OpenRepository(worktree.Path())
	checkFatal(t, err)
	worktreeFromRepo, err := worktreeRepo.Worktrees.OpenFromRepository()
	checkFatal(t, err)

	if worktreeFromRepo.Name() != worktree.Name() {
		t.Fatalf("wrong name (expected %s, got %s)", worktreeFromRepo.Name(), worktree.Name())
	}
}

func TestWorktreeIsPrunable(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)

	worktreeTemporaryPath, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)
	defer func() { checkFatal(t, os.RemoveAll(worktreeTemporaryPath)) }()

	worktree, err := repo.Worktrees.Add("testWorktree", filepath.Join(worktreeTemporaryPath, "worktree"), nil)
	checkFatal(t, err)
	err = worktree.Lock("test")
	checkFatal(t, err)

	isPrunableWithoutLockedFlag, err := worktree.IsPrunable(WorktreePruneValid)
	checkFatal(t, err)
	if isPrunableWithoutLockedFlag {
		t.Fatal("worktree shouldn't be prunable without the WorktreePruneLocked flag")
	}
	isPrunableWithLockedFlag, err := worktree.IsPrunable(WorktreePruneValid | WorktreePruneLocked)
	checkFatal(t, err)
	if !isPrunableWithLockedFlag {
		t.Fatal("worktree should be prunable with the WorktreePruneLocked flag")
	}

	err = worktree.Prune(WorktreePruneValid | WorktreePruneLocked)
	checkFatal(t, err)
}

func TestWorktreeCanBeLockedAndUnlocked(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)

	worktreeTemporaryPath, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)
	defer func() { checkFatal(t, os.RemoveAll(worktreeTemporaryPath)) }()

	worktree, err := repo.Worktrees.Add("testWorktree", filepath.Join(worktreeTemporaryPath, "worktree"), nil)
	checkFatal(t, err)
	notLocked, err := worktree.Unlock()
	checkFatal(t, err)
	if !notLocked {
		t.Fatal("worktree should be unlocked by default")
	}

	expectedReason := "toTestIt"
	err = worktree.Lock(expectedReason)
	checkFatal(t, err)
	isLocked, reason, err := worktree.IsLocked()
	checkFatal(t, err)
	if !isLocked {
		t.Fatal("worktree should be locked after the locking operation")
	}
	if expectedReason != reason {
		t.Fatalf("locked reason doesn't match: %s != %s", expectedReason, reason)
	}

	notLocked, err = worktree.Unlock()
	checkFatal(t, err)
	if notLocked {
		t.Fatal("worktree was lock before so notLocked should be false")
	}
	isLocked, _, err = worktree.IsLocked()
	checkFatal(t, err)
	if isLocked {
		t.Fatal("worktree should be unlocked after the Unlock() call")
	}
}
