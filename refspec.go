package git

/*
#include <git2.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type Refspec struct {
	doNotCompare
	ptr *C.git_refspec
}

// ParseRefspec parses a given refspec string
func ParseRefspec(input string, isFetch bool) (*Refspec, error) {
	var ptr *C.git_refspec

	cinput := C.CString(input)
	defer C.free(unsafe.Pointer(cinput))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_refspec_parse(&ptr, cinput, cbool(isFetch))
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	spec := &Refspec{ptr: ptr}
	runtime.SetFinalizer(spec, (*Refspec).Free)
	return spec, nil
}

// Free releases a refspec object which has been created by ParseRefspec
func (s *Refspec) Free() {
	runtime.SetFinalizer(s, nil)
	C.git_refspec_free(s.ptr)
}

// Direction returns the refspec's direction
func (s *Refspec) Direction() ConnectDirection {
	direction := C.git_refspec_direction(s.ptr)
	return ConnectDirection(direction)
}

// Src returns the refspec's source specifier
func (s *Refspec) Src() string {
	var ret string
	cstr := C.git_refspec_src(s.ptr)

	if cstr != nil {
		ret = C.GoString(cstr)
	}

	runtime.KeepAlive(s)
	return ret
}

// Dst returns the refspec's destination specifier
func (s *Refspec) Dst() string {
	var ret string
	cstr := C.git_refspec_dst(s.ptr)

	if cstr != nil {
		ret = C.GoString(cstr)
	}

	runtime.KeepAlive(s)
	return ret
}

// Force returns the refspec's force-update setting
func (s *Refspec) Force() bool {
	force := C.git_refspec_force(s.ptr)
	return force != 0
}

// String returns the refspec's string representation
func (s *Refspec) String() string {
	var ret string
	cstr := C.git_refspec_string(s.ptr)

	if cstr != nil {
		ret = C.GoString(cstr)
	}

	runtime.KeepAlive(s)
	return ret
}

// SrcMatches checks if a refspec's source descriptor matches a reference
func (s *Refspec) SrcMatches(refname string) bool {
	cname := C.CString(refname)
	defer C.free(unsafe.Pointer(cname))

	matches := C.git_refspec_src_matches(s.ptr, cname)
	return matches != 0
}

// SrcMatches checks if a refspec's destination descriptor matches a reference
func (s *Refspec) DstMatches(refname string) bool {
	cname := C.CString(refname)
	defer C.free(unsafe.Pointer(cname))

	matches := C.git_refspec_dst_matches(s.ptr, cname)
	return matches != 0
}

// Transform a reference to its target following the refspec's rules
func (s *Refspec) Transform(refname string) (string, error) {
	buf := C.git_buf{}

	cname := C.CString(refname)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_refspec_transform(&buf, s.ptr, cname)
	if ret < 0 {
		return "", MakeGitError(ret)
	}
	defer C.git_buf_dispose(&buf)

	return C.GoString(buf.ptr), nil
}

// Rtransform converts a target reference to its source reference following the
// refspec's rules
func (s *Refspec) Rtransform(refname string) (string, error) {
	buf := C.git_buf{}

	cname := C.CString(refname)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_refspec_rtransform(&buf, s.ptr, cname)
	if ret < 0 {
		return "", MakeGitError(ret)
	}
	defer C.git_buf_dispose(&buf)

	return C.GoString(buf.ptr), nil
}
