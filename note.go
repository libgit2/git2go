package git

/*
#include <git2.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// Note
type Note struct {
	ptr *C.git_note
}

// Free frees a git_note object
func (n *Note) Free() error {
	if n.ptr == nil {
		return ErrInvalid
	}
	runtime.SetFinalizer(n, nil)
	C.git_note_free(n.ptr)
	n.ptr = nil
	return nil
}

// Author returns the signature of the note author
func (n *Note) Author() *Signature {
	ptr := C.git_note_author(n.ptr)
	return newSignatureFromC(ptr)
}

// Id returns the note object's id
func (n *Note) Id() *Oid {
	ptr := C.git_note_id(n.ptr)
	return newOidFromC(ptr)
}

// Committer returns the signature of the note committer
func (n *Note) Committer() *Signature {
	ptr := C.git_note_committer(n.ptr)
	return newSignatureFromC(ptr)
}

// Message returns the note message
func (n *Note) Message() string {
	return C.GoString(C.git_note_message(n.ptr))
}

// NoteIterator
type NoteIterator struct {
	ptr *C.git_note_iterator
}

// NewNoteIterator creates a new iterator for notes
func (repo *Repository) NewNoteIterator(ref string) (*NoteIterator, error) {
	var cref *C.char
	if ref == "" {
		cref = nil
	} else {
		cref = C.CString(ref)
		defer C.free(unsafe.Pointer(cref))
	}

	var ptr *C.git_note_iterator

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if ret := C.git_note_iterator_new(&ptr, repo.ptr, cref); ret < 0 {
		return nil, MakeGitError(ret)
	}

	iter := &NoteIterator{ptr: ptr}
	runtime.SetFinalizer(iter, (*NoteIterator).Free)
	return iter, nil
}

// Free frees the note interator
func (v *NoteIterator) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_note_iterator_free(v.ptr)
}

// Next returns the current item (note id & annotated id) and advances the
// iterator internally to the next item
func (it *NoteIterator) Next() (noteId, annotatedId *Oid, err error) {
	noteId, annotatedId = new(Oid), new(Oid)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if ret := C.git_note_next(noteId.toC(), annotatedId.toC(), it.ptr); ret < 0 {
		err = MakeGitError(ret)
	}
	return
}
