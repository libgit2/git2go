package git

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func allReflogEntries(t *testing.T, repo *Repository, refName string) (entries []*ReflogEntry) {
	rl, err := repo.ReadReflog(refName)
	checkFatal(t, err)
	defer rl.Free()

	for i := uint(0); i < rl.EntryCount(); i++ {
		entries = append(entries, rl.EntryByIndex(i))
	}
	return entries
}

// assertEntriesEqual will assert that the reflogs match with the exception of
// the signature time (it is not reliably deterministic to predict the
// signature time during many reference updates)
func assertEntriesEqual(t *testing.T, got, want []*ReflogEntry) {
	if len(got) != len(want) {
		t.Fatalf("got %d length, wanted %d length", len(got), len(want))
	}

	for i := 0; i < len(got); i++ {
		gi := got[i]
		wi := want[i]
		// remove the signature time to make the results deterministic
		gi.Committer.When = time.Time{}
		wi.Committer.When = time.Time{}
		// check committer separately to print results clearly
		if !reflect.DeepEqual(gi.Committer, wi.Committer) {
			t.Fatalf("got committer %v, want committer %v",
				gi.Committer, wi.Committer)
		}
		if !reflect.DeepEqual(gi, wi) {
			t.Fatalf("got %v, want %v", gi, wi)
		}
	}
}

func TestReflog(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitID, treeId := seedTestRepo(t, repo)

	testRefName := "refs/heads/test"

	// configure committer for deterministic reflog entries
	cfg, err := repo.Config()
	checkFatal(t, err)

	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
	}

	checkFatal(t, cfg.SetString("user.name", sig.Name))
	checkFatal(t, cfg.SetString("user.email", sig.Email))

	checkFatal(t, repo.References.EnsureLog(testRefName))
	_, err = repo.References.Create(testRefName, commitID, true, "first update")
	checkFatal(t, err)
	got := allReflogEntries(t, repo, testRefName)
	want := []*ReflogEntry{
		&ReflogEntry{
			New:       commitID,
			Old:       &Oid{},
			Committer: sig,
			Message:   "first update",
		},
	}

	// create additional commits and verify they are added to reflog
	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)
	for i := 0; i < 10; i++ {
		nextEntry := &ReflogEntry{
			Old:       commitID,
			Committer: sig,
			Message:   fmt.Sprintf("commit: %d", i),
		}

		commit, err := repo.LookupCommit(commitID)
		checkFatal(t, err)

		commitID, err = repo.CreateCommit(testRefName, sig, sig, fmt.Sprint(i), tree, commit)
		checkFatal(t, err)

		nextEntry.New = commitID

		want = append([]*ReflogEntry{nextEntry}, want...)
	}

	t.Run("ReadReflog", func(t *testing.T) {
		got = allReflogEntries(t, repo, testRefName)
		assertEntriesEqual(t, got, want)
	})

	t.Run("DropEntry", func(t *testing.T) {
		rl, err := repo.ReadReflog(testRefName)
		checkFatal(t, err)
		defer rl.Free()

		gotBefore := allReflogEntries(t, repo, testRefName)

		checkFatal(t, rl.DropEntry(0, false))
		checkFatal(t, rl.Write())

		gotAfter := allReflogEntries(t, repo, testRefName)

		assertEntriesEqual(t, gotAfter, gotBefore[1:])
	})

	t.Run("AppendEntry", func(t *testing.T) {
		logs := allReflogEntries(t, repo, testRefName)

		rl, err := repo.ReadReflog(testRefName)
		checkFatal(t, err)
		defer rl.Free()

		newOID := NewOidFromBytes([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})
		checkFatal(t, rl.AppendEntry(newOID, sig, "synthetic"))
		checkFatal(t, rl.Write())

		want := append([]*ReflogEntry{
			&ReflogEntry{
				New:       newOID,
				Old:       logs[0].New,
				Committer: sig,
				Message:   "synthetic",
			},
		}, logs...)
		got := allReflogEntries(t, repo, testRefName)
		assertEntriesEqual(t, got, want)
	})

	t.Run("RenameReflog", func(t *testing.T) {
		logs := allReflogEntries(t, repo, testRefName)
		newRefName := "refs/heads/new"

		checkFatal(t, repo.RenameReflog(testRefName, newRefName))
		assertEntriesEqual(t, allReflogEntries(t, repo, testRefName), nil)
		assertEntriesEqual(t, allReflogEntries(t, repo, newRefName), logs)

		checkFatal(t, repo.RenameReflog(newRefName, testRefName))
		assertEntriesEqual(t, allReflogEntries(t, repo, testRefName), logs)
		assertEntriesEqual(t, allReflogEntries(t, repo, newRefName), nil)
	})

	t.Run("DeleteReflog", func(t *testing.T) {
		checkFatal(t, repo.DeleteReflog(testRefName))
		assertEntriesEqual(t, allReflogEntries(t, repo, testRefName), nil)
	})

}
