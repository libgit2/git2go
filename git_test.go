package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	ret := m.Run()

	// Ensure that we are not leaking any pointer handles.
	pointerHandles.Lock()
	if len(pointerHandles.handles) > 0 {
		for h, ptr := range pointerHandles.handles {
			fmt.Printf("%016p: %v %+v\n", h, reflect.TypeOf(ptr), ptr)
		}
		panic("pointer handle list not empty")
	}
	pointerHandles.Unlock()

	Shutdown()

	os.Exit(ret)
}

func cleanupTestRepo(t *testing.T, r *Repository) {
	var err error
	if r.IsBare() {
		err = os.RemoveAll(r.Path())
	} else {
		err = os.RemoveAll(r.Workdir())
	}
	checkFatal(t, err)

	r.Free()
}

func createTestRepo(t *testing.T) *Repository {
	// figure out where we can create the test repo
	path, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)
	repo, err := InitRepository(path, false)
	checkFatal(t, err)

	tmpfile := "README"
	err = ioutil.WriteFile(path+"/"+tmpfile, []byte("foo\n"), 0644)

	checkFatal(t, err)

	return repo
}

func createBareTestRepo(t *testing.T) *Repository {
	// figure out where we can create the test repo
	path, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)
	repo, err := InitRepository(path, true)
	checkFatal(t, err)

	return repo
}

// commitOptions contains any extra options for creating commits in the seed repo
type commitOptions struct {
	CommitSigningCallback
}

func seedTestRepo(t *testing.T, repo *Repository) (*Oid, *Oid) {
	return seedTestRepoOpt(t, repo, commitOptions{})
}

func seedTestRepoOpt(t *testing.T, repo *Repository, opts commitOptions) (*Oid, *Oid) {
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
	err = idx.Write()
	checkFatal(t, err)
	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)
	commitId, err := repo.CreateCommit("HEAD", sig, sig, message, tree)
	checkFatal(t, err)

	if opts.CommitSigningCallback != nil {
		commit, err := repo.LookupCommit(commitId)
		checkFatal(t, err)

		signature, signatureField, err := opts.CommitSigningCallback(commit.ContentToSign())
		checkFatal(t, err)

		oid, err := commit.WithSignature(signature, signatureField)
		checkFatal(t, err)
		newCommit, err := repo.LookupCommit(oid)
		checkFatal(t, err)
		head, err := repo.Head()
		checkFatal(t, err)
		_, err = repo.References.Create(
			head.Name(),
			newCommit.Id(),
			true,
			"repoint to signed commit",
		)
		checkFatal(t, err)
	}

	return commitId, treeId
}

func pathInRepo(repo *Repository, name string) string {
	return path.Join(path.Dir(path.Dir(repo.Path())), name)
}

func updateReadme(t *testing.T, repo *Repository, content string) (*Oid, *Oid) {
	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	tmpfile := "README"
	err = ioutil.WriteFile(pathInRepo(repo, tmpfile), []byte(content), 0644)
	checkFatal(t, err)

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("README")
	checkFatal(t, err)
	err = idx.Write()
	checkFatal(t, err)
	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	currentBranch, err := repo.Head()
	checkFatal(t, err)
	currentTip, err := repo.LookupCommit(currentBranch.Target())
	checkFatal(t, err)

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)
	commitId, err := repo.CreateCommit("HEAD", sig, sig, message, tree, currentTip)
	checkFatal(t, err)

	return commitId, treeId
}

func TestOidZero(t *testing.T) {
	t.Parallel()
	var zeroId Oid

	if !zeroId.IsZero() {
		t.Error("Zero Oid is not zero")
	}
}

func TestEmptyOid(t *testing.T) {
	t.Parallel()
	_, err := NewOid("")
	if err == nil || !IsErrorCode(err, ErrorCodeGeneric) {
		t.Fatal("Should have returned invalid error")
	}
}
