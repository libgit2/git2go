// +build static,!system_libgit2

package git

/*
#cgo windows CFLAGS: -I${SRCDIR}/static-build/install/include/
#cgo windows LDFLAGS: -L${SRCDIR}/static-build/install/lib/ -lgit2 -lwinhttp -lws2_32 -lole32 -lrpcrt4 -lcrypt32
#cgo !windows pkg-config: --static ${SRCDIR}/static-build/install/lib/pkgconfig/libgit2.pc
#cgo CFLAGS: -DLIBGIT2_STATIC
#include <git2.h>

#if LIBGIT2_VER_MAJOR != 1 || LIBGIT2_VER_MINOR < 1 || LIBGIT2_VER_MINOR > 2
# error "Invalid libgit2 version; this git2go supports libgit2 between v1.1.0 and v1.2.0"
#endif
*/
import "C"
