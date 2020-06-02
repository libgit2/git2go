package git

/*
#include <git2.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

// Trailer represents a single git message trailer.
type Trailer struct {
	Key   string
	Value string
}

// MessageTrailers parses trailers out of a message, returning a slice of
// Trailer structs. Trailers are key/value pairs in the last paragraph of a
// message, not including any patches or conflicts that may be present.
func MessageTrailers(message string) ([]Trailer, error) {
	var trailersC C.git_message_trailer_array

	messageC := C.CString(message)
	defer C.free(unsafe.Pointer(messageC))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_message_trailers(&trailersC, messageC)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}
	defer C.git_message_trailer_array_free(&trailersC)
	trailers := make([]Trailer, trailersC.count)
	var trailer *C.git_message_trailer
	for i, p := 0, uintptr(unsafe.Pointer(trailersC.trailers)); i < int(trailersC.count); i, p = i+1, p+unsafe.Sizeof(C.git_message_trailer{}) {
		trailer = (*C.git_message_trailer)(unsafe.Pointer(p))
		trailers[i] = Trailer{Key: C.GoString(trailer.key), Value: C.GoString(trailer.value)}
	}
	return trailers, nil
}
