package git

/*
#include <git2.h>
#include <git2/errors.h>

#include "wrap.h"
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type Config struct {
	ptr *C.git_config
}

func (c *Config) LookupInt32(name string) (int32, error) {
	var out C.int32_t
	var err *C.git_error
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if ret := C.e_git_config_get_int32(&out, c.ptr, cname, &err); ret < 0 {
		return 0, makeError(ret, err)
	}

	return int32(out), nil
}

func (c *Config) LookupInt64(name string) (int64, error) {
	var out C.int64_t
	var err *C.git_error
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if ret := C.e_git_config_get_int64(&out, c.ptr, cname, &err); ret < 0 {
		return 0, makeError(ret, err)
	}

	return int64(out), nil
}

func (c *Config) LookupString(name string) (string, error) {
	var ptr *C.char
	var err *C.git_error
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if ret := C.e_git_config_get_string(&ptr, c.ptr, cname, &err); ret < 0 {
		return "", makeError(ret, err)
	}

	return C.GoString(ptr), nil
}

func (c *Config) Set(name, value string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	var err *C.git_error
	ret := C.e_git_config_set_string(c.ptr, cname, cvalue, &err)
	return makeError(ret, err)
}

func (c *Config) Free() {
	runtime.SetFinalizer(c, nil)
	C.git_config_free(c.ptr)
}
