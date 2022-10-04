//go:build !static
// +build !static

package git

/*
#cgo pkg-config: libgit2
#cgo CFLAGS: -DLIBGIT2_DYNAMIC
#include <git2.h>

#if LIBGIT2_VER_MAJOR != 1 || LIBGIT2_VER_MINOR < 5 || LIBGIT2_VER_MINOR > 5
# error "Invalid libgit2 version; this git2go supports libgit2 between v1.5.0 and v1.5.0"
#endif
*/
import "C"
