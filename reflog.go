package git

/*
#include <git2.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

// Reflog is a log of changes for a reference
type Reflog struct {
	ptr  *C.git_reflog
	repo *Repository
	name string
}

func newRefLogFromC(ptr *C.git_reflog, repo *Repository, name string) *Reflog {
	l := &Reflog{
		ptr:  ptr,
		repo: repo,
		name: name,
	}
	runtime.SetFinalizer(l, (*Reflog).Free)
	return l
}

func (repo *Repository) ReadReflog(name string) (*Reflog, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var ptr *C.git_reflog

	ecode := C.git_reflog_read(&ptr, repo.ptr, cname)
	runtime.KeepAlive(repo)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newRefLogFromC(ptr, repo, name), nil
}

func (repo *Repository) DeleteReflog(name string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ecode := C.git_reflog_delete(repo.ptr, cname)
	runtime.KeepAlive(repo)
	if ecode < 0 {
		return MakeGitError(ecode)
	}

	return nil
}

func (repo *Repository) RenameReflog(oldName, newName string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cOldName := C.CString(oldName)
	defer C.free(unsafe.Pointer(cOldName))

	cNewName := C.CString(newName)
	defer C.free(unsafe.Pointer(cNewName))

	ecode := C.git_reflog_rename(repo.ptr, cOldName, cNewName)
	runtime.KeepAlive(repo)
	if ecode < 0 {
		return MakeGitError(ecode)
	}

	return nil
}

func (l *Reflog) Write() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_reflog_write(l.ptr)
	runtime.KeepAlive(l)
	if ecode < 0 {
		return MakeGitError(ecode)
	}
	return nil
}

func (l *Reflog) EntryCount() uint {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	count := C.git_reflog_entrycount(l.ptr)
	runtime.KeepAlive(l)
	return uint(count)
}

// ReflogEntry specifies a reference change
type ReflogEntry struct {
	Old       *Oid
	New       *Oid
	Committer *Signature
	Message   string // may be empty
}

func newReflogEntry(entry *C.git_reflog_entry) *ReflogEntry {
	return &ReflogEntry{
		New:       newOidFromC(C.git_reflog_entry_id_new(entry)),
		Old:       newOidFromC(C.git_reflog_entry_id_old(entry)),
		Committer: newSignatureFromC(C.git_reflog_entry_committer(entry)),
		Message:   C.GoString(C.git_reflog_entry_message(entry)),
	}
}

func (l *Reflog) EntryByIndex(index uint) *ReflogEntry {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	entry := C.git_reflog_entry_byindex(l.ptr, C.size_t(index))
	if entry == nil {
		return nil
	}

	goEntry := newReflogEntry(entry)
	runtime.KeepAlive(l)

	return goEntry
}

func (l *Reflog) DropEntry(index uint, rewriteHistory bool) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var rewriteHistoryInt int
	if rewriteHistory {
		rewriteHistoryInt = 1
	}

	ecode := C.git_reflog_drop(l.ptr, C.size_t(index), C.int(rewriteHistoryInt))
	runtime.KeepAlive(l)
	if ecode < 0 {
		return MakeGitError(ecode)
	}

	return nil
}

func (l *Reflog) AppendEntry(oid *Oid, committer *Signature, message string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cSignature, err := committer.toC()
	if err != nil {
		return err
	}
	defer C.git_signature_free(cSignature)

	cMsg := C.CString(message)
	defer C.free(unsafe.Pointer(cMsg))

	C.git_reflog_append(l.ptr, oid.toC(), cSignature, cMsg)
	runtime.KeepAlive(l)

	return nil
}

func (l *Reflog) Free() {
	runtime.SetFinalizer(l, nil)
	C.git_reflog_free(l.ptr)
}
