package git

import (
	"io/ioutil"
	"testing"
)

func checkout(t *testing.T, repo *Repository, commit *Commit) {
	tree, err := commit.Tree()
	if err != nil {
		t.Fatal(err)
	}

	err = repo.CheckoutTree(tree, &CheckoutOpts{Strategy: CheckoutSafe})
	if err != nil {
		t.Fatal(err)
	}

	err = repo.SetHeadDetached(commit.Id(), commit.Author(), "checkout")
	if err != nil {
		t.Fatal(err)
	}
}

const content = "Herro, Worrd!"

func readReadme(t *testing.T, repo *Repository) string {
	bytes, err := ioutil.ReadFile(pathInRepo(repo, "README"))
	if err != nil {
		t.Fatal(err)
	}
	return string(bytes)
}

func TestCherrypick(t *testing.T) {
	repo := createTestRepo(t)
	c1, _ := seedTestRepo(t, repo)
	c2, _ := updateReadme(t, repo, content)

	commit1, err := repo.LookupCommit(c1)
	if err != nil {
		t.Fatal(err)
	}
	commit2, err := repo.LookupCommit(c2)
	if err != nil {
		t.Fatal(err)
	}

	checkout(t, repo, commit1)

	if readReadme(t, repo) == content {
		t.Fatalf("README has wrong content after checking out initial commit")
	}

	opts, err := DefaultCherrypickOptions()
	if err != nil {
		t.Fatal(err)
	}

	err = repo.Cherrypick(commit2, opts)
	if err != nil {
		t.Fatal(err)
	}

	if readReadme(t, repo) != content {
		t.Fatalf("README has wrong contents after cherry-picking")
	}

	state := repo.State()
	if state != RepositoryStateCherrypick {
		t.Fatal("Incorrect repository state: ", state)
	}

	err = repo.StateCleanup()
	if err != nil {
		t.Fatal(err)
	}

	state = repo.State()
	if state != RepositoryStateNone {
		t.Fatal("Incorrect repository state: ", state)
	}
}
