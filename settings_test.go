package git

import (
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
