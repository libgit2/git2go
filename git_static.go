// +build static

package git

/*
#cgo CFLAGS: -I${SRCDIR}/vendor/libgit2/include
#cgo LDFLAGS: -L${SRCDIR}/vendor/libgit2/build/ -lgit2
#cgo windows LDFLAGS: -lwinhttp
#cgo !windows pkg-config: --static ${SRCDIR}/vendor/libgit2/build/libgit2.pc
#include <git2.h>

#if LIBGIT2_VER_MAJOR != 0 || LIBGIT2_VER_MINOR != 26
# error "Invalid libgit2 version; this git2go supports libgit2 v0.26"
#endif

*/
import "C"
