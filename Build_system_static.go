//go:build static && system_libgit2
// +build static,system_libgit2

package git

/*
#cgo pkg-config: libgit2 --static
#cgo CFLAGS: -DLIBGIT2_STATIC
#include <git2.h>

#if LIBGIT2_VER_MAJOR != 1 || LIBGIT2_VER_MINOR < 6 || LIBGIT2_VER_MINOR > 6
# error "Invalid libgit2 version; this git2go supports libgit2 v1.6.x"
#endif
*/
import "C"
