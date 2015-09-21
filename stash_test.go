package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestStash(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	prepareStashRepo(t, repo)

	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Now(),
	}

	stash1, err := repo.Stashes.Save(sig, "First stash", StashDefault)
	checkFatal(t, err)

	_, err = repo.LookupCommit(stash1)
	checkFatal(t, err)

	b, err := ioutil.ReadFile(pathInRepo(repo, "README"))
	checkFatal(t, err)
	if string(b) == "Update README goes to stash\n" {
		t.Errorf("README still contains the uncommitted changes")
	}

	if !fileExistsInRepo(repo, "untracked.txt") {
		t.Errorf("untracked.txt doesn't exist in the repo; should be untracked")
	}

	// Apply: default

	opts, err := DefaultStashApplyOptions()
	checkFatal(t, err)

	err = repo.Stashes.Apply(0, opts)
	checkFatal(t, err)

	b, err = ioutil.ReadFile(pathInRepo(repo, "README"))
	checkFatal(t, err)
	if string(b) != "Update README goes to stash\n" {
		t.Errorf("README changes aren't here")
	}

	// Apply: no stash for the given index

	err = repo.Stashes.Apply(1, opts)
	if !IsErrorCode(err, ErrNotFound) {
		t.Errorf("expecting GIT_ENOTFOUND error code %d, got %v", ErrNotFound, err)
	}

	// Apply: callback stopped

	opts.ProgressCallback = func(progress StashApplyProgress) error {
		if progress == StashApplyProgressCheckoutModified {
			return fmt.Errorf("Stop")
		}
		return nil
	}

	err = repo.Stashes.Apply(0, opts)
	if err.Error() != "Stop" {
		t.Errorf("expecting error 'Stop', got %v", err)
	}

	// Create second stash with ignored files

	os.MkdirAll(pathInRepo(repo, "tmp"), os.ModeDir|os.ModePerm)
	err = ioutil.WriteFile(pathInRepo(repo, "tmp/ignored.txt"), []byte("Ignore me\n"), 0644)
	checkFatal(t, err)

	stash2, err := repo.Stashes.Save(sig, "Second stash", StashIncludeIgnored)
	checkFatal(t, err)

	if fileExistsInRepo(repo, "tmp/ignored.txt") {
		t.Errorf("tmp/ignored.txt should not exist anymore in the work dir")
	}

	// Stash foreach

	expected := []stash{
		{0, "On master: Second stash", stash2.String()},
		{1, "On master: First stash", stash1.String()},
	}
	checkStashes(t, repo, expected)

	// Stash pop

	opts, _ = DefaultStashApplyOptions()
	err = repo.Stashes.Pop(1, opts)
	checkFatal(t, err)

	b, err = ioutil.ReadFile(pathInRepo(repo, "README"))
	checkFatal(t, err)
	if string(b) != "Update README goes to stash\n" {
		t.Errorf("README changes aren't here")
	}

	expected = []stash{
		{0, "On master: Second stash", stash2.String()},
	}
	checkStashes(t, repo, expected)

	// Stash drop

	err = repo.Stashes.Drop(0)
	checkFatal(t, err)

	expected = []stash{}
	checkStashes(t, repo, expected)
}

type stash struct {
	index int
	msg   string
	id    string
}

func checkStashes(t *testing.T, repo *Repository, expected []stash) {
	var actual []stash

	repo.Stashes.Foreach(func(index int, msg string, id *Oid) error {
		stash := stash{index, msg, id.String()}
		if len(expected) > len(actual) {
			if s := expected[len(actual)]; s.id == "" {
				stash.id = "" //  don't check id
			}
		}
		actual = append(actual, stash)
		return nil
	})

	if len(expected) > 0 && !reflect.DeepEqual(expected, actual) {
		// The failure happens at wherever we were called, not here
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			t.Fatalf("Unable to get caller")
		}
		t.Errorf("%v:%v: expecting %#v\ngot %#v", path.Base(file), line, expected, actual)
	}
}

func prepareStashRepo(t *testing.T, repo *Repository) {
	seedTestRepo(t, repo)

	err := ioutil.WriteFile(pathInRepo(repo, ".gitignore"), []byte("tmp\n"), 0644)
	checkFatal(t, err)

	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Now(),
	}

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath(".gitignore")
	checkFatal(t, err)
	treeID, err := idx.WriteTree()
	checkFatal(t, err)
	err = idx.Write()
	checkFatal(t, err)

	currentBranch, err := repo.Head()
	checkFatal(t, err)
	currentTip, err := repo.LookupCommit(currentBranch.Target())
	checkFatal(t, err)

	message := "Add .gitignore\n"
	tree, err := repo.LookupTree(treeID)
	checkFatal(t, err)
	_, err = repo.CreateCommit("HEAD", sig, sig, message, tree, currentTip)
	checkFatal(t, err)

	err = ioutil.WriteFile(pathInRepo(repo, "README"), []byte("Update README goes to stash\n"), 0644)
	checkFatal(t, err)

	err = ioutil.WriteFile(pathInRepo(repo, "untracked.txt"), []byte("Hello, World\n"), 0644)
	checkFatal(t, err)
}

func fileExistsInRepo(repo *Repository, name string) bool {
	if _, err := os.Stat(pathInRepo(repo, name)); err != nil {
		return false
	}
	return true
}
