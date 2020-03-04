// +build !static

package git

/*
#include <git2.h>
#cgo pkg-config: libgit2

#if LIBGIT2_VER_MAJOR != 0 || LIBGIT2_VER_MINOR != 99
# error "Invalid libgit2 version; this git2go supports libgit2 v0.99"
#endif

*/
import "C"
