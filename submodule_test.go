package git

import (
	"testing"
)

func TestSubmoduleForeach(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	_, err := repo.AddSubmodule("http://example.org/submodule", "submodule", true)
	checkFatal(t, err)

	i := 0
	err = repo.ForeachSubmodule(func(sub *Submodule, name string) int {
		i++
		return 0
	})
	checkFatal(t, err)

	if i != 1 {
		t.Fatalf("expected one submodule found but got %d", i)
	}
}
