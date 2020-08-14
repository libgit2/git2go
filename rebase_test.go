package git

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

// Tests

func TestRebaseAbort(t *testing.T) {
	// TEST DATA

	// Inputs
	branchName := "emile"
	masterCommit := "something"
	emileCommits := []string{
		"fou",
		"barre",
	}

	// Outputs
	expectedHistory := []string{
		"Test rebase, Baby! " + emileCommits[1],
		"Test rebase, Baby! " + emileCommits[0],
		"This is a commit\n",
	}

	// TEST
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)

	// Setup a repo with 2 branches and a different tree
	err := setupRepoForRebase(repo, masterCommit, branchName)
	checkFatal(t, err)

	// Create several commits in emile
	for _, commit := range emileCommits {
		_, err = commitSomething(repo, commit, commit)
		checkFatal(t, err)
	}

	// Check history
	actualHistory, err := commitMsgsList(repo)
	checkFatal(t, err)
	assertStringList(t, expectedHistory, actualHistory)

	// Rebase onto master
	rebase, err := performRebaseOnto(repo, "master", nil)
	checkFatal(t, err)
	defer rebase.Free()

	// Abort rebase
	rebase.Abort()

	// Check history is still the same
	actualHistory, err = commitMsgsList(repo)
	checkFatal(t, err)
	assertStringList(t, expectedHistory, actualHistory)
}

func TestRebaseNoConflicts(t *testing.T) {
	// TEST DATA

	// Inputs
	branchName := "emile"
	masterCommit := "something"
	emileCommits := []string{
		"fou",
		"barre",
		"ouich",
	}

	// Outputs
	expectedHistory := []string{
		"Test rebase, Baby! " + emileCommits[2],
		"Test rebase, Baby! " + emileCommits[1],
		"Test rebase, Baby! " + emileCommits[0],
		"Test rebase, Baby! " + masterCommit,
		"This is a commit\n",
	}

	// TEST
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepo(t, repo)

	// Try to open existing rebase
	oRebase, err := repo.OpenRebase(nil)
	if err == nil {
		t.Fatal("Did not expect to find a rebase in progress")
	}

	// Setup a repo with 2 branches and a different tree
	err = setupRepoForRebase(repo, masterCommit, branchName)
	checkFatal(t, err)

	// Create several commits in emile
	for _, commit := range emileCommits {
		_, err = commitSomething(repo, commit, commit)
		checkFatal(t, err)
	}

	// Rebase onto master
	rebase, err := performRebaseOnto(repo, "master", nil)
	checkFatal(t, err)
	defer rebase.Free()

	// Open existing rebase
	oRebase, err = repo.OpenRebase(nil)
	checkFatal(t, err)
	defer oRebase.Free()
	if oRebase == nil {
		t.Fatal("Expected to find an existing rebase in progress")
	}

	// Finish the rebase properly
	err = rebase.Finish()
	checkFatal(t, err)

	// Check no more rebase is in progress
	oRebase, err = repo.OpenRebase(nil)
	if err == nil {
		t.Fatal("Did not expect to find a rebase in progress")
	}

	// Check history is in correct order
	actualHistory, err := commitMsgsList(repo)
	checkFatal(t, err)
	assertStringList(t, expectedHistory, actualHistory)
}

func TestRebaseGpgSigned(t *testing.T) {
	// TEST DATA

	entity, err := openpgp.NewEntity("Namey mcnameface", "test comment", "test@example.com", nil)
	checkFatal(t, err)

	opts, err := DefaultRebaseOptions()
	checkFatal(t, err)

	signCommitContent := func(commitContent string) (string, string, error) {
		cipherText := new(bytes.Buffer)
		err := openpgp.ArmoredDetachSignText(cipherText, entity, strings.NewReader(commitContent), &packet.Config{})
		if err != nil {
			return "", "", errors.New("error signing payload")
		}

		return cipherText.String(), "", nil
	}
	opts.SigningCallback = signCommitContent

	commitOpts := commitOpts{
		CommitSigningCb: signCommitContent,
	}

	// Inputs
	branchName := "emile"
	masterCommit := "something"
	emileCommits := []string{
		"fou",
		"barre",
		"ouich",
	}

	// Outputs
	expectedHistory := []string{
		"Test rebase, Baby! " + emileCommits[2],
		"Test rebase, Baby! " + emileCommits[1],
		"Test rebase, Baby! " + emileCommits[0],
		"Test rebase, Baby! " + masterCommit,
		"This is a commit\n",
	}

	// TEST
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)
	seedTestRepoOpt(t, repo, commitOpts)

	// Try to open existing rebase
	_, err = repo.OpenRebase(nil)
	if err == nil {
		t.Fatal("Did not expect to find a rebase in progress")
	}

	// Setup a repo with 2 branches and a different tree
	err = setupRepoForRebaseOpts(repo, masterCommit, branchName, commitOpts)
	checkFatal(t, err)

	// Create several commits in emile
	for _, commit := range emileCommits {
		_, err = commitSomethingOpts(repo, commit, commit, commitOpts)
		checkFatal(t, err)
	}

	// Rebase onto master
	rebase, err := performRebaseOnto(repo, "master", &opts)
	checkFatal(t, err)
	defer rebase.Free()

	// Finish the rebase properly
	err = rebase.Finish()
	checkFatal(t, err)

	// Check history is in correct order
	actualHistory, err := commitMsgsList(repo)
	checkFatal(t, err)
	assertStringList(t, expectedHistory, actualHistory)

	checkAllCommitsSigned(t, entity, repo)
}

func checkAllCommitsSigned(t *testing.T, entity *openpgp.Entity, repo *Repository) {
	head, err := headCommit(repo)
	checkFatal(t, err)
	defer head.Free()

	parent := head

	err = checkCommitSigned(t, entity, parent)
	checkFatal(t, err)

	for parent.ParentCount() != 0 {
		parent = parent.Parent(0)
		defer parent.Free()

		err = checkCommitSigned(t, entity, parent)
		checkFatal(t, err)
	}
}

func checkCommitSigned(t *testing.T, entity *openpgp.Entity, commit *Commit) error {
	t.Helper()

	signature, signedData, err := commit.ExtractSignature()
	if err != nil {
		t.Logf("No signature on commit\n%s", commit.ContentToSign())
		return err
	}

	_, err = openpgp.CheckArmoredDetachedSignature(openpgp.EntityList{entity}, strings.NewReader(signedData), bytes.NewBufferString(signature))
	if err != nil {
		t.Logf("Commit is not signed correctly\n%s", commit.ContentToSign())
		return err
	}

	return nil
}

// Utils
func setupRepoForRebase(repo *Repository, masterCommit, branchName string) error {
	return setupRepoForRebaseOpts(repo, masterCommit, branchName, commitOpts{})
}

func setupRepoForRebaseOpts(repo *Repository, masterCommit, branchName string, opts commitOpts) error {
	// Create a new branch from master
	err := createBranch(repo, branchName)
	if err != nil {
		return err
	}

	// Create a commit in master
	_, err = commitSomethingOpts(repo, masterCommit, masterCommit, opts)
	if err != nil {
		return err
	}

	// Switch to emile
	err = repo.SetHead("refs/heads/" + branchName)
	if err != nil {
		return err
	}

	// Check master commit is not in emile branch
	if entryExists(repo, masterCommit) {
		return errors.New(masterCommit + " entry should not exist in " + branchName + " branch.")
	}

	return nil
}

func performRebaseOnto(repo *Repository, branch string, opts *RebaseOptions) (*Rebase, error) {
	master, err := repo.LookupBranch(branch, BranchLocal)
	if err != nil {
		return nil, err
	}
	defer master.Free()

	onto, err := repo.AnnotatedCommitFromRef(master.Reference)
	if err != nil {
		return nil, err
	}
	defer onto.Free()

	// Init rebase
	rebase, err := repo.InitRebase(nil, nil, onto, opts)
	if err != nil {
		return nil, err
	}

	// Check no operation has been started yet
	rebaseOperationIndex, err := rebase.CurrentOperationIndex()
	if rebaseOperationIndex != RebaseNoOperation && err != ErrRebaseNoOperation {
		return nil, errors.New("No operation should have been started yet")
	}

	// Iterate in rebase operations regarding operation count
	opCount := int(rebase.OperationCount())
	for op := 0; op < opCount; op++ {
		operation, err := rebase.Next()
		if err != nil {
			return nil, err
		}

		// Check operation index is correct
		rebaseOperationIndex, err = rebase.CurrentOperationIndex()
		if int(rebaseOperationIndex) != op {
			return nil, errors.New("Bad operation index")
		}
		if !operationsAreEqual(rebase.OperationAt(uint(op)), operation) {
			return nil, errors.New("Rebase operations should be equal")
		}

		// Get current rebase operation created commit
		commit, err := repo.LookupCommit(operation.Id)
		if err != nil {
			return nil, err
		}
		defer commit.Free()

		// Apply commit
		err = rebase.Commit(operation.Id, signature(), signature(), commit.Message())
		if err != nil {
			return nil, err
		}
	}

	return rebase, nil
}

func operationsAreEqual(l, r *RebaseOperation) bool {
	return l.Exec == r.Exec && l.Type == r.Type && l.Id.String() == r.Id.String()
}

func createBranch(repo *Repository, branch string) error {
	commit, err := headCommit(repo)
	if err != nil {
		return err
	}
	defer commit.Free()
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

func headCommit(repo *Repository) (*Commit, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	defer head.Free()

	commit, err := repo.LookupCommit(head.Target())
	if err != nil {
		return nil, err
	}

	return commit, nil
}

func headTree(repo *Repository) (*Tree, error) {
	headCommit, err := headCommit(repo)
	if err != nil {
		return nil, err
	}
	defer headCommit.Free()

	tree, err := headCommit.Tree()
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func commitSomething(repo *Repository, something, content string) (*Oid, error) {
	return commitSomethingOpts(repo, something, content, commitOpts{})
}

func commitSomethingOpts(repo *Repository, something, content string, commitOpts commitOpts) (*Oid, error) {
	headCommit, err := headCommit(repo)
	if err != nil {
		return nil, err
	}
	defer headCommit.Free()

	index, err := NewIndex()
	if err != nil {
		return nil, err
	}
	defer index.Free()

	blobOID, err := repo.CreateBlobFromBuffer([]byte(content))
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
	defer newTree.Free()

	commit, err := repo.CreateCommit("HEAD", signature(), signature(), "Test rebase, Baby! "+something, newTree, headCommit)
	if err != nil {
		return nil, err
	}

	if commitOpts.CommitSigningCb != nil {
		commit, err := repo.LookupCommit(commit)
		if err != nil {
			return nil, err
		}

		oid, err := commit.WithSignatureUsing(commitOpts.CommitSigningCb)
		if err != nil {
			return nil, err
		}
		newCommit, err := repo.LookupCommit(oid)
		if err != nil {
			return nil, err
		}
		head, err := repo.Head()
		if err != nil {
			return nil, err
		}
		_, err = repo.References.Create(
			head.Name(),
			newCommit.Id(),
			true,
			"repoint to signed commit",
		)
		if err != nil {
			return nil, err
		}
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
	headTree, err := headTree(repo)
	if err != nil {
		return false
	}
	defer headTree.Free()

	_, err = headTree.EntryByPath(file)

	return err == nil
}

func commitMsgsList(repo *Repository) ([]string, error) {
	head, err := headCommit(repo)
	if err != nil {
		return nil, err
	}
	defer head.Free()

	var commits []string

	parent := head.Parent(0)
	defer parent.Free()
	commits = append(commits, head.Message(), parent.Message())

	for parent.ParentCount() != 0 {
		parent = parent.Parent(0)
		defer parent.Free()
		commits = append(commits, parent.Message())
	}

	return commits, nil
}

func assertStringList(t *testing.T, expected, actual []string) {
	if len(expected) != len(actual) {
		t.Fatal("Lists are not the same size, expected " + strconv.Itoa(len(expected)) +
			", got " + strconv.Itoa(len(actual)))
	}
	for index, element := range expected {
		if element != actual[index] {
			t.Error("Expected element " + strconv.Itoa(index) + " to be " + element + ", got " + actual[index])
		}
	}
}
