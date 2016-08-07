package git

import (
	"testing"
	"time"
)

func createBranch(repo *Repository, branch string) error {
	head, err := repo.Head()
	if err != nil {
		return err
	}
	commit, err := repo.LookupCommit(head.Target())
	if err != nil {
		return err
	}
	_, err = repo.CreateBranch(branch, commit, false)
	if err != nil {
		return err
	}

	return nil
}

func signature() *Signature {
	return &Signature{
		Name:  "Emile",
		Email: "emile@emile.com",
		When:  time.Now(),
	}
}

func commitSomething(repo *Repository, something string) (*Oid, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	headCommit, err := repo.LookupCommit(head.Target())
	if err != nil {
		return nil, err
	}

	index, err := NewIndex()
	if err != nil {
		return nil, err
	}
	defer index.Free()

	blobOID, err := repo.CreateBlobFromBuffer([]byte("fou"))
	if err != nil {
		return nil, err
	}

	entry := &IndexEntry{
		Mode: FilemodeBlob,
		Id:   blobOID,
		Path: something,
	}

	if err := index.Add(entry); err != nil {
		return nil, err
	}

	newTreeOID, err := index.WriteTreeTo(repo)
	if err != nil {
		return nil, err
	}

	newTree, err := repo.LookupTree(newTreeOID)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	commit, err := repo.CreateCommit("HEAD", signature(), signature(), "Test rebase onto, Baby! "+something, newTree, headCommit)
	if err != nil {
		return nil, err
	}

	opts := &CheckoutOpts{
		Strategy: CheckoutRemoveUntracked | CheckoutForce,
	}
	err = repo.CheckoutIndex(index, opts)
	if err != nil {
		return nil, err
	}

	return commit, nil
}

func entryExists(repo *Repository, file string) bool {
	head, err := repo.Head()
	if err != nil {
		return false
	}
	headCommit, err := repo.LookupCommit(head.Target())
	if err != nil {
		return false
	}
	headTree, err := headCommit.Tree()
	if err != nil {
		return false
	}
	_, err = headTree.EntryByPath(file)

	return err == nil
}

func TestRebaseOnto(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	fileInMaster := "something"
	fileInEmile := "something else"

	// Seed master
	seedTestRepo(t, repo)

	// Create a new branch from master
	err := createBranch(repo, "emile")
	checkFatal(t, err)

	// Create a commit in master
	_, err = commitSomething(repo, fileInMaster)
	checkFatal(t, err)

	// Switch to this emile
	err = repo.SetHead("refs/heads/emile")
	checkFatal(t, err)

	// Check master commit is not in emile branch
	if entryExists(repo, fileInMaster) {
		t.Fatal("something entry should not exist in emile branch")
	}

	// Create a commit in emile
	_, err = commitSomething(repo, fileInEmile)
	checkFatal(t, err)

	// Rebase onto master
	master, err := repo.LookupBranch("master", BranchLocal)
	branch, err := repo.AnnotatedCommitFromRef(master.Reference)
	checkFatal(t, err)

	rebase, err := repo.RebaseInit(nil, nil, branch, nil)
	checkFatal(t, err)
	defer rebase.Free()

	operation, err := rebase.Next()
	checkFatal(t, err)

	commit, err := repo.LookupCommit(operation.ID)
	checkFatal(t, err)

	err = rebase.Commit(operation.ID, signature(), signature(), commit.Message())
	checkFatal(t, err)

	rebase.Finish()

	// Check master commit is now also in emile branch
	if !entryExists(repo, fileInMaster) {
		t.Fatal("something entry should now exist in emile branch")
	}
}
