package git

/*
#include <git2.h>
#include <git2/errors.h>
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
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_config_get_int32(&out, c.ptr, cname)
	if ret < 0 {
		return 0, LastError()
	}

	return int32(out), nil
}

func (c *Config) LookupInt64(name string) (int64, error) {
	var out C.int64_t
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_config_get_int64(&out, c.ptr, cname)
	if ret < 0 {
		return 0, LastError()
	}

	return int64(out), nil
}

func (c *Config) LookupString(name string) (string, error) {
	var ptr *C.char
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_config_get_string(&ptr, c.ptr, cname)
	if ret < 0 {
		return "", LastError()
	}

	return C.GoString(ptr), nil
}


func (c *Config) LookupBool(name string) (bool, error) {
	var out C.int
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ret := C.git_config_get_bool(&out, c.ptr, cname)
	if ret < 0 {
		return false, LastError()
	}

	return out != 0, nil
}

func (c *Config) SetString(name, value string) (err error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_config_set_string(c.ptr, cname, cvalue)
	if ret < 0 {
		return LastError()
	}

	return nil
}

func (c *Config) Free() {
	runtime.SetFinalizer(c, nil)
	C.git_config_free(c.ptr)
}

func (c *Config) SetInt32(name string, value int32) (err error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ret := C.git_config_set_int32(c.ptr, cname, C.int32_t(value))
	if ret < 0 {
		return LastError()
	}

	return nil
}

func (c *Config) SetInt64(name string, value int64) (err error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ret := C.git_config_set_int64(c.ptr, cname, C.int64_t(value))
	if ret < 0 {
		return LastError()
	}

	return nil
}

func (c *Config) SetBool(name string, value bool) (err error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ret := C.git_config_set_bool(c.ptr, cname, cbool(value))
	if ret < 0 {
		return LastError()
	}

	return nil
}

func (c *Config) SetMultivar(name, regexp, value string) (err error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cregexp := C.CString(regexp)
	defer C.free(unsafe.Pointer(cregexp))

	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	ret := C.git_config_set_multivar(c.ptr, cname, cregexp, cvalue)
	if ret < 0 {
		return LastError()
	}

	return nil
}

func (c *Config) Delete(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ret := C.git_config_delete_entry(c.ptr, cname)

	if ret < 0 {
		return LastError()
	}

	return nil
}
