package git

import (
	"testing"
)

func TestSubmoduleForeach(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	_, err := repo.Submodules.Add("http://example.org/submodule", "submodule", true)
	checkFatal(t, err)

	i := 0
	err = repo.Submodules.Foreach(func(sub *Submodule, name string) error {
		i++
		return nil
	})
	checkFatal(t, err)

	if i != 1 {
		t.Fatalf("expected one submodule found but got %d", i)
	}
}
