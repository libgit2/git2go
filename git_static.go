// +build static

package git

/*
#cgo CFLAGS: -I${SRCDIR}/vendor/libgit2/include
#cgo LDFLAGS: -L${SRCDIR}/vendor/libgit2/build/ -lgit2
#cgo windows LDFLAGS: -lwinhttp
#cgo !windows pkg-config: --static ${SRCDIR}/vendor/libgit2/build/libgit2.pc
*/
import "C"
