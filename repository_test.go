package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenBareRepository(t *testing.T) {
	t.Parallel()

	repo := createBareTestRepo(t)
	defer cleanupTestRepo(t, repo)

	opened, err := OpenBareRepository(repo.Path())
	checkFatal(t, err)
	defer opened.Free()

	if !opened.IsBare() {
		t.Fatal("expected opened repository to be bare")
	}

	if opened.Workdir() != "" {
		t.Fatalf("expected bare repository to have empty workdir, got %q", opened.Workdir())
	}
}

func TestOpenBareRepositoryFromGitDir(t *testing.T) {
	t.Parallel()

	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	opened, err := OpenBareRepository(repo.Path())
	checkFatal(t, err)
	defer opened.Free()

	if !opened.IsBare() {
		t.Fatal("expected repository opened from .git dir to be bare")
	}
}

func TestOpenBareRepositoryNotFound(t *testing.T) {
	t.Parallel()

	missingPath := filepath.Join(os.TempDir(), "git2go-does-not-exist-open-bare")
	opened, err := OpenBareRepository(missingPath)

	if opened != nil {
		t.Fatal("expected nil repository for missing path")
	}

	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestCreateCommitBuffer(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

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

	for encoding, expected := range map[MessageEncoding]string{
		MessageEncodingUTF8: `tree b7119b11e8ef7a1a5a34d3ac87f5b075228ac81e
author Rand Om Hacker <random@hacker.com> 1362576600 +0100
committer Rand Om Hacker <random@hacker.com> 1362576600 +0100

This is a commit
`,
		MessageEncoding("ASCII"): `tree b7119b11e8ef7a1a5a34d3ac87f5b075228ac81e
author Rand Om Hacker <random@hacker.com> 1362576600 +0100
committer Rand Om Hacker <random@hacker.com> 1362576600 +0100
encoding ASCII

This is a commit
`,
	} {
		encoding := encoding
		expected := expected
		t.Run(string(encoding), func(t *testing.T) {
			buf, err := repo.CreateCommitBuffer(sig, sig, encoding, message, tree)
			checkFatal(t, err)

			if expected != string(buf) {
				t.Errorf("mismatched commit buffer, expected %v, got %v", expected, string(buf))
			}
		})
	}
}

func TestCreateCommitFromIds(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

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
	expectedCommitId, err := repo.CreateCommit("HEAD", sig, sig, message, tree)
	checkFatal(t, err)

	commitId, err := repo.CreateCommitFromIds("", sig, sig, message, treeId)
	checkFatal(t, err)

	if !expectedCommitId.Equal(commitId) {
		t.Errorf("mismatched commit ids, expected %v, got %v", expectedCommitId.String(), commitId.String())
	}
}

func TestRepositorySetConfig(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

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

	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)
	_, err = repo.CreateCommit("HEAD", sig, sig, message, tree)
	checkFatal(t, err)

	repoConfig, err := repo.Config()
	checkFatal(t, err)

	temp := Config{}
	localConfig, err := temp.OpenLevel(repoConfig, ConfigLevelLocal)
	checkFatal(t, err)
	repoConfig = nil

	err = repo.SetConfig(localConfig)
	checkFatal(t, err)

	configFieldName := "core.filemode"
	err = localConfig.SetBool(configFieldName, true)
	checkFatal(t, err)

	localConfig = nil

	repoConfig, err = repo.Config()
	checkFatal(t, err)

	result, err := repoConfig.LookupBool(configFieldName)
	checkFatal(t, err)
	if result != true {
		t.Fatal("result must be true")
	}
}

func TestRepositoryItemPath(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	gitDir, err := repo.ItemPath(RepositoryItemGitDir)
	checkFatal(t, err)
	if gitDir == "" {
		t.Error("expected not empty gitDir")
	}
}

func TestCommitParents(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	parents, err := repo.CommitParents()
	checkFatal(t, err)

	if len(parents) != 1 {
		t.Fatalf("expected 1 parent, got %d", len(parents))
	}

	// The parent should be HEAD
	head, err := repo.Head()
	checkFatal(t, err)
	defer head.Free()

	if !parents[0].Id().Equal(head.Target()) {
		t.Fatalf("expected parent to be HEAD (%s), got %s", head.Target(), parents[0].Id())
	}

	for _, p := range parents {
		p.Free()
	}
}
