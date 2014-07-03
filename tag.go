package git

/*
#include <git2.h>
*/
import "C"

// Tag
type Tag struct {
	gitObject
	cast_ptr *C.git_tag
}

func (t Tag) Message() string {
	return C.GoString(C.git_tag_message(t.cast_ptr))
}

func (t Tag) Name() string {
	return C.GoString(C.git_tag_name(t.cast_ptr))
}

func (t Tag) Tagger() *Signature {
	cast_ptr := C.git_tag_tagger(t.cast_ptr)
	return newSignatureFromC(cast_ptr)
}

func (t Tag) Target() Object {
	var ptr *C.git_object
	ret := C.git_tag_target(&ptr, t.cast_ptr)

	if ret != 0 {
		return nil
	}

	return allocObject(ptr, t.repo)
}

func (t Tag) TargetId() *Oid {
	return newOidFromC(C.git_tag_target_id(t.cast_ptr))
}

func (t Tag) TargetType() ObjectType {
	return ObjectType(C.git_tag_target_type(t.cast_ptr))
}
