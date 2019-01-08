// +build static

package git

/*
#cgo windows CFLAGS: -I${SRCDIR}/static-build/install/include/
#cgo windows LDFLAGS: -L${SRCDIR}/static-build/install/lib/ -lgit2 -lwinhttp
#cgo !windows pkg-config: --static ${SRCDIR}/static-build/install/lib/pkgconfig/libgit2.pc
#include <git2.h>

#if LIBGIT2_VER_MAJOR != 0 || LIBGIT2_VER_MINOR != 27
# error "Invalid libgit2 version; this git2go supports libgit2 v0.27"
#endif

*/
import "C"
