//go:build static && system_libgit2
// +build static,system_libgit2

package git

/*
#cgo pkg-config: libgit2 --static
#cgo CFLAGS: -DLIBGIT2_STATIC
#include <git2.h>

#if LIBGIT2_VER_MAJOR != 1 || LIBGIT2_VER_MINOR < 3 || LIBGIT2_VER_MINOR > 3
# error "Invalid libgit2 version; this git2go supports libgit2 between v1.3.0 and v1.3.0"
#endif
*/
import "C"
