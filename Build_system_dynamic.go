//go:build !static
// +build !static

package git

/*
#cgo pkg-config: libgit2
#cgo CFLAGS: -DLIBGIT2_DYNAMIC
#include <git2.h>

#if LIBGIT2_VER_MAJOR != 1 || LIBGIT2_VER_MINOR < 9 || LIBGIT2_VER_MINOR > 9
# error "Invalid libgit2 version; this git2go supports libgit2 between v1.9.0 and v1.9.x"
#endif
*/
import "C"
