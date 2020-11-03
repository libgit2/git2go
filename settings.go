package git

/*
#include <git2.h>

int _go_git_opts_get_search_path(int level, git_buf *buf)
{
    return git_libgit2_opts(GIT_OPT_GET_SEARCH_PATH, level, buf);
}

int _go_git_opts_set_search_path(int level, const char *path)
{
    return git_libgit2_opts(GIT_OPT_SET_SEARCH_PATH, level, path);
}

int _go_git_opts_set_size_t(int opt, size_t val)
{
    return git_libgit2_opts(opt, val);
}

int _go_git_opts_set_cache_object_limit(git_object_t type, size_t size)
{
    return git_libgit2_opts(GIT_OPT_SET_CACHE_OBJECT_LIMIT, type, size);
}

int _go_git_opts_get_size_t(int opt, size_t *val)
{
    return git_libgit2_opts(opt, val);
}

int _go_git_opts_get_size_t_size_t(int opt, size_t *val1, size_t *val2)
{
    return git_libgit2_opts(opt, val1, val2);
}
*/
import "C"
import (
	"runtime"
	"unsafe"
)

func SearchPath(level ConfigLevel) (string, error) {
	var buf C.git_buf
	defer C.git_buf_dispose(&buf)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C._go_git_opts_get_search_path(C.int(level), &buf)
	if err < 0 {
		return "", MakeGitError(err)
	}

	return C.GoString(buf.ptr), nil
}

func SetSearchPath(level ConfigLevel, path string) error {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C._go_git_opts_set_search_path(C.int(level), cpath)
	if err < 0 {
		return MakeGitError(err)
	}

	return nil
}

func MwindowSize() (int, error) {
	return getSizet(C.GIT_OPT_GET_MWINDOW_SIZE)
}

func SetMwindowSize(size int) error {
	return setSizet(C.GIT_OPT_SET_MWINDOW_SIZE, size)
}

func MwindowMappedLimit() (int, error) {
	return getSizet(C.GIT_OPT_GET_MWINDOW_MAPPED_LIMIT)
}

func SetMwindowMappedLimit(size int) error {
	return setSizet(C.GIT_OPT_SET_MWINDOW_MAPPED_LIMIT, size)
}

func EnableCaching(enabled bool) error {
	if enabled {
		return setSizet(C.GIT_OPT_ENABLE_CACHING, 1)
	} else {
		return setSizet(C.GIT_OPT_ENABLE_CACHING, 0)
	}
}

func EnableStrictHashVerification(enabled bool) error {
	if enabled {
		return setSizet(C.GIT_OPT_ENABLE_STRICT_HASH_VERIFICATION, 1)
	} else {
		return setSizet(C.GIT_OPT_ENABLE_STRICT_HASH_VERIFICATION, 0)
	}
}

func CachedMemory() (current int, allowed int, err error) {
	return getSizetSizet(C.GIT_OPT_GET_CACHED_MEMORY)
}

// deprecated: You should use `CachedMemory()` instead.
func GetCachedMemory() (current int, allowed int, err error) {
	return CachedMemory()
}

func SetCacheMaxSize(maxSize int) error {
	return setSizet(C.GIT_OPT_SET_CACHE_MAX_SIZE, maxSize)
}

func SetCacheObjectLimit(objectType ObjectType, size int) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C._go_git_opts_set_cache_object_limit(C.git_object_t(objectType), C.size_t(size))
	if err < 0 {
		return MakeGitError(err)
	}

	return nil
}

func getSizet(opt C.int) (int, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var val C.size_t
	err := C._go_git_opts_get_size_t(opt, &val)
	if err < 0 {
		return 0, MakeGitError(err)
	}

	return int(val), nil
}

func getSizetSizet(opt C.int) (int, int, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var val1, val2 C.size_t
	err := C._go_git_opts_get_size_t_size_t(opt, &val1, &val2)
	if err < 0 {
		return 0, 0, MakeGitError(err)
	}

	return int(val1), int(val2), nil
}

func setSizet(opt C.int, val int) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cval := C.size_t(val)
	err := C._go_git_opts_set_size_t(opt, cval)
	if err < 0 {
		return MakeGitError(err)
	}

	return nil
}
