package git

import (
	"testing"
)

func TestDefaultSignatureFromEnv(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	// Configure user for the repo
	cfg, err := repo.Config()
	checkFatal(t, err)
	defer cfg.Free()
	err = cfg.SetString("user.name", "Test Author")
	checkFatal(t, err)
	err = cfg.SetString("user.email", "author@example.com")
	checkFatal(t, err)

	author, committer, err := repo.DefaultSignatureFromEnv()
	checkFatal(t, err)

	if author == nil {
		t.Fatal("expected non-nil author")
	}
	if committer == nil {
		t.Fatal("expected non-nil committer")
	}

	if author.Name != "Test Author" {
		t.Errorf("expected author name 'Test Author', got %q", author.Name)
	}
	if author.Email != "author@example.com" {
		t.Errorf("expected author email 'author@example.com', got %q", author.Email)
	}
}
