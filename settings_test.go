package git

import (
	"io/ioutil"
	"os"
	"testing"
)

type pathPair struct {
	Level ConfigLevel
	Path  string
}

func TestSearchPath(t *testing.T) {
	paths := []pathPair{
		pathPair{ConfigLevelSystem, "/tmp/system"},
		pathPair{ConfigLevelGlobal, "/tmp/global"},
		pathPair{ConfigLevelXDG, "/tmp/xdg"},
	}

	for _, pair := range paths {
		err := SetSearchPath(pair.Level, pair.Path)
		checkFatal(t, err)

		actual, err := SearchPath(pair.Level)
		checkFatal(t, err)

		if pair.Path != actual {
			t.Fatal("Search paths don't match")
		}
	}
}

func TestMmapSizes(t *testing.T) {
	size := 42 * 1024

	err := SetMwindowSize(size)
	checkFatal(t, err)

	actual, err := MwindowSize()
	if size != actual {
		t.Fatal("Sizes don't match")
	}

	err = SetMwindowMappedLimit(size)
	checkFatal(t, err)

	actual, err = MwindowMappedLimit()
	if size != actual {
		t.Fatal("Sizes don't match")
	}
}

func TestEnableCaching(t *testing.T) {
	err := EnableCaching(false)
	checkFatal(t, err)

	err = EnableCaching(true)
	checkFatal(t, err)
}

func TestEnableStrictHashVerification(t *testing.T) {
	err := EnableStrictHashVerification(false)
	checkFatal(t, err)

	err = EnableStrictHashVerification(true)
	checkFatal(t, err)
}

func TestEnableFsyncGitDir(t *testing.T) {
	err := EnableFsyncGitDir(false)
	checkFatal(t, err)

	err = EnableFsyncGitDir(true)
	checkFatal(t, err)
}

func TestCachedMemory(t *testing.T) {
	current, allowed, err := CachedMemory()
	checkFatal(t, err)

	if current < 0 {
		t.Fatal("current < 0")
	}

	if allowed < 0 {
		t.Fatal("allowed < 0")
	}
}

func TestSetCacheMaxSize(t *testing.T) {
	err := SetCacheMaxSize(0)
	checkFatal(t, err)

	err = SetCacheMaxSize(1024 * 1024)
	checkFatal(t, err)

	// revert to default 256MB
	err = SetCacheMaxSize(256 * 1024 * 1024)
	checkFatal(t, err)
}

func TestHomeDir(t *testing.T) {
	t.Parallel()
	dir, err := HomeDir()
	checkFatal(t, err)

	if dir == "" {
		t.Fatal("HomeDir returned empty string")
	}
}

func TestSetHomeDir(t *testing.T) {
	t.Parallel()
	original, err := HomeDir()
	checkFatal(t, err)

	tmpDir, err := ioutil.TempDir("", "git2go-homedir")
	checkFatal(t, err)
	defer os.RemoveAll(tmpDir)

	err = SetHomeDir(tmpDir)
	checkFatal(t, err)

	actual, err := HomeDir()
	checkFatal(t, err)
	if actual != tmpDir {
		t.Fatalf("expected %q, got %q", tmpDir, actual)
	}

	// Restore original
	err = SetHomeDir(original)
	checkFatal(t, err)
}

func TestServerConnectTimeout(t *testing.T) {
	t.Parallel()
	err := SetServerConnectTimeout(5000)
	checkFatal(t, err)

	val, err := ServerConnectTimeout()
	checkFatal(t, err)
	if val != 5000 {
		t.Fatalf("expected 5000, got %d", val)
	}

	// Reset to default
	err = SetServerConnectTimeout(0)
	checkFatal(t, err)
}

func TestServerTimeout(t *testing.T) {
	t.Parallel()
	err := SetServerTimeout(10000)
	checkFatal(t, err)

	val, err := ServerTimeout()
	checkFatal(t, err)
	if val != 10000 {
		t.Fatalf("expected 10000, got %d", val)
	}

	// Reset to default
	err = SetServerTimeout(0)
	checkFatal(t, err)
}
