package git

import (
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	config, err := repo.Config()
	checkFatal(t, err)

	_, err = config.LookupInt32("core.repositoryformatversion")
	checkFatal(t, err)
	_, err = config.LookupString("this.doesnt.exist")
	if err == nil {
		t.Fatal("No error returned")
	}
	gitErr, ok := err.(*GitError)
	if !ok {
		t.Fatal("Bad error type")
	}
	if gitErr.Code != -3 {
		t.Fatalf("Expected ENOTFOUND got %v\n", gitErr.Code)
	}
}
