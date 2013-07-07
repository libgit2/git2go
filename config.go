package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
*/
import "C"
import (
	"unsafe"
)

type Config struct {
	ptr *C.git_config
}

func (c *Config) LookupInt32(name string) (v int32, err error) {
	var out C.int32_t
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ret := C.git_config_get_int32(&out, c.ptr, cname)
	if ret < 0 {
		return 0, MakeGitError(ret)
	}

	return int32(out), nil
}

func (c *Config) LookupInt64(name string) (v int64, err error) {
	var out C.int64_t
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ret := C.git_config_get_int64(&out, c.ptr, cname)
	if ret < 0 {
		return 0, MakeGitError(ret)
	}

	return int64(out), nil
}

func (c *Config) LookupString(name string) (v string, err error) {
	var ptr *C.char
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ret := C.git_config_get_string(&ptr, c.ptr, cname)
	if ret < 0 {
		return "", MakeGitError(ret)
	}

	return C.GoString(ptr), nil
}

func (c *Config) Set(name, value string) (err error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	ret := C.git_config_set_string(c.ptr, cname, cvalue)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}
