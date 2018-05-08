// +build static

package git

/*
#cgo windows CFLAGS: -I${SRCDIR}/vendor/libgit2/include
#cgo windows LDFLAGS: -L${SRCDIR}/vendor/libgit2/build/ -lgit2 -lwinhttp
#cgo !windows pkg-config: --static ${SRCDIR}/vendor/libgit2/build/libgit2.pc
#include <git2.h>

#if LIBGIT2_VER_MAJOR != 0 || LIBGIT2_VER_MINOR != 27
# error "Invalid libgit2 version; this git2go supports libgit2 v0.27"
#endif

*/
import "C"
