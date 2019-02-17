// +build !static

package git

/*
#cgo pkg-config: libgit2
#cgo CFLAGS: -DLIBGIT2_DYNAMIC
#include <git2.h>

#if LIBGIT2_VER_MAJOR != 1 || LIBGIT2_VER_MINOR != 0
# error "Invalid libgit2 version; this git2go supports libgit2 v1.0"
#endif

*/
import "C"
