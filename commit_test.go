package git

import (
	"testing"
)

func TestCreateCommitFromStage(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	// Configure user for the repo
	cfg, err := repo.Config()
	checkFatal(t, err)
	defer cfg.Free()
	err = cfg.SetString("user.name", "Test User")
	checkFatal(t, err)
	err = cfg.SetString("user.email", "test@example.com")
	checkFatal(t, err)

	// Stage a file
	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("README")
	checkFatal(t, err)
	err = idx.Write()
	checkFatal(t, err)

	// Create commit from stage
	oid, err := repo.CreateCommitFromStage("initial commit from stage", nil)
	checkFatal(t, err)

	if oid == nil || oid.IsZero() {
		t.Fatal("expected a valid commit OID")
	}

	// Verify the commit
	commit, err := repo.LookupCommit(oid)
	checkFatal(t, err)
	defer commit.Free()

	if commit.Message() != "initial commit from stage" {
		t.Fatalf("expected 'initial commit from stage', got %q", commit.Message())
	}
}

func TestCreateCommitFromStageAllowEmpty(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	cfg, err := repo.Config()
	checkFatal(t, err)
	defer cfg.Free()
	err = cfg.SetString("user.name", "Test User")
	checkFatal(t, err)
	err = cfg.SetString("user.email", "test@example.com")
	checkFatal(t, err)

	seedTestRepo(t, repo)

	// Try empty commit without AllowEmptyCommit — should fail
	_, err = repo.CreateCommitFromStage("empty commit", nil)
	if err == nil {
		t.Fatal("expected error for empty commit without AllowEmptyCommit")
	}

	// Try with AllowEmptyCommit
	oid, err := repo.CreateCommitFromStage("empty commit", &CommitCreateOptions{
		AllowEmptyCommit: true,
	})
	checkFatal(t, err)
	if oid == nil || oid.IsZero() {
		t.Fatal("expected a valid commit OID for empty commit")
	}
}
