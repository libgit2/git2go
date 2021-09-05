package git

/*
#include <git2.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// This object represents the possible operations which can be
// performed on the collection of notes for a repository.
type NoteCollection struct {
	doNotCompare
	repo *Repository
}

// Create adds a note for an object
func (c *NoteCollection) Create(
	ref string, author, committer *Signature, id *Oid,
	note string, force bool) (*Oid, error) {

	oid := new(Oid)

	var cref *C.char
	if ref == "" {
		cref = nil
	} else {
		cref = C.CString(ref)
		defer C.free(unsafe.Pointer(cref))
	}

	authorSig, err := author.toC()
	if err != nil {
		return nil, err
	}
	defer C.git_signature_free(authorSig)

	committerSig, err := committer.toC()
	if err != nil {
		return nil, err
	}
	defer C.git_signature_free(committerSig)

	cnote := C.CString(note)
	defer C.free(unsafe.Pointer(cnote))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_note_create(
		oid.toC(), c.repo.ptr, cref, authorSig,
		committerSig, id.toC(), cnote, cbool(force))
	runtime.KeepAlive(c)
	runtime.KeepAlive(id)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return oid, nil
}

// Read reads the note for an object
func (c *NoteCollection) Read(ref string, id *Oid) (*Note, error) {
	var cref *C.char
	if ref == "" {
		cref = nil
	} else {
		cref = C.CString(ref)
		defer C.free(unsafe.Pointer(cref))
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_note
	ret := C.git_note_read(&ptr, c.repo.ptr, cref, id.toC())
	runtime.KeepAlive(c)
	runtime.KeepAlive(id)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newNoteFromC(ptr, c.repo), nil
}

// Remove removes the note for an object
func (c *NoteCollection) Remove(ref string, author, committer *Signature, id *Oid) error {
	var cref *C.char
	if ref == "" {
		cref = nil
	} else {
		cref = C.CString(ref)
		defer C.free(unsafe.Pointer(cref))
	}

	authorSig, err := author.toC()
	if err != nil {
		return err
	}
	defer C.git_signature_free(authorSig)

	committerSig, err := committer.toC()
	if err != nil {
		return err
	}
	defer C.git_signature_free(committerSig)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_note_remove(c.repo.ptr, cref, authorSig, committerSig, id.toC())
	runtime.KeepAlive(c)
	runtime.KeepAlive(id)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

// DefaultRef returns the default notes reference for a repository
func (c *NoteCollection) DefaultRef() (string, error) {
	buf := C.git_buf{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_note_default_ref(&buf, c.repo.ptr)
	runtime.KeepAlive(c)
	if ecode < 0 {
		return "", MakeGitError(ecode)
	}

	ret := C.GoString(buf.ptr)
	C.git_buf_dispose(&buf)

	return ret, nil
}

// Note
type Note struct {
	doNotCompare
	ptr *C.git_note
	r   *Repository
}

func newNoteFromC(ptr *C.git_note, r *Repository) *Note {
	note := &Note{ptr: ptr, r: r}
	runtime.SetFinalizer(note, (*Note).Free)
	return note
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
	runtime.KeepAlive(n)
	return newOidFromC(ptr)
}

// Committer returns the signature of the note committer
func (n *Note) Committer() *Signature {
	ptr := C.git_note_committer(n.ptr)
	runtime.KeepAlive(n)
	return newSignatureFromC(ptr)
}

// Message returns the note message
func (n *Note) Message() string {
	ret := C.GoString(C.git_note_message(n.ptr))
	runtime.KeepAlive(n)
	return ret
}

// NoteIterator
type NoteIterator struct {
	doNotCompare
	ptr *C.git_note_iterator
	r   *Repository
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

	ret := C.git_note_iterator_new(&ptr, repo.ptr, cref)
	runtime.KeepAlive(repo)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	iter := &NoteIterator{ptr: ptr, r: repo}
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

	ret := C.git_note_next(noteId.toC(), annotatedId.toC(), it.ptr)
	runtime.KeepAlive(noteId)
	runtime.KeepAlive(annotatedId)
	runtime.KeepAlive(it)
	if ret < 0 {
		err = MakeGitError(ret)
	}
	return
}
