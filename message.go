package git

/*
#include <git2.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type Trailer struct {
	key   string
	value string
}

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
	for i, p := 0, uintptr(unsafe.Pointer(trailersC.trailers)); i < int(trailersC.count); p += unsafe.Sizeof(C.git_message_trailer{}) {
		trailer = (*C.git_message_trailer)(unsafe.Pointer(p))
		trailers[i] = Trailer{key: C.GoString(trailer.key), value: C.GoString(trailer.value)}
		i++
	}
	return trailers, nil
}
