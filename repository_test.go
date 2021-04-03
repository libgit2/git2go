package git

import (
	"testing"
	"time"
)

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