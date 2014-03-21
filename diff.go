package git

/*
#include <git2.h>

extern int _go_git_diff_foreach(git_diff *diff, int eachFile, int eachHunk, int eachLine, void *payload);
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type DiffFlag int

const (
	DiffFlagBinary    DiffFlag = C.GIT_DIFF_FLAG_BINARY
	DiffFlagNotBinary          = C.GIT_DIFF_FLAG_NOT_BINARY
	DiffFlagValidOid           = C.GIT_DIFF_FLAG_VALID_OID
)

type Delta int

const (
	DeltaUnmodified Delta = C.GIT_DELTA_UNMODIFIED
	DeltaAdded            = C.GIT_DELTA_ADDED
	DeltaDeleted          = C.GIT_DELTA_DELETED
	DeltaModified         = C.GIT_DELTA_MODIFIED
	DeltaRenamed          = C.GIT_DELTA_RENAMED
	DeltaCopied           = C.GIT_DELTA_COPIED
	DeltaIgnored          = C.GIT_DELTA_IGNORED
	DeltaUntracked        = C.GIT_DELTA_UNTRACKED
	DeltaTypeChange       = C.GIT_DELTA_TYPECHANGE
)

type DiffLineType int

const (
	DiffLineContext      DiffLineType = C.GIT_DIFF_LINE_CONTEXT
	DiffLineAddition                  = C.GIT_DIFF_LINE_ADDITION
	DiffLineDeletion                  = C.GIT_DIFF_LINE_DELETION
	DiffLineContextEOFNL              = C.GIT_DIFF_LINE_CONTEXT_EOFNL
	DiffLineAddEOFNL                  = C.GIT_DIFF_LINE_ADD_EOFNL
	DiffLineDelEOFNL                  = C.GIT_DIFF_LINE_DEL_EOFNL

	DiffLineFileHdr = C.GIT_DIFF_LINE_FILE_HDR
	DiffLineHunkHdr = C.GIT_DIFF_LINE_HUNK_HDR
	DiffLineBinary  = C.GIT_DIFF_LINE_BINARY
)

type DiffFile struct {
	Path  string
	Oid   *Oid
	Size  int
	Flags DiffFlag
	Mode  uint16
}

func newDiffFileFromC(file *C.git_diff_file) *DiffFile {
	return &DiffFile{
		Path:  C.GoString(file.path),
		Oid:   newOidFromC(&file.oid),
		Size:  int(file.size),
		Flags: DiffFlag(file.flags),
		Mode:  uint16(file.mode),
	}
}

type DiffDelta struct {
	Status     Delta
	Flags      DiffFlag
	Similarity uint16
	OldFile    *DiffFile
	NewFile    *DiffFile
}

func newDiffDeltaFromC(delta *C.git_diff_delta) *DiffDelta {
	return &DiffDelta{
		Status:     Delta(delta.status),
		Flags:      DiffFlag(delta.flags),
		Similarity: uint16(delta.similarity),
		OldFile:    newDiffFileFromC(&delta.old_file),
		NewFile:    newDiffFileFromC(&delta.new_file),
	}
}

type DiffHunk struct {
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Header   string
	DiffDelta
}

func newDiffHunkFromC(delta *C.git_diff_delta, hunk *C.git_diff_hunk) *DiffHunk {
	return &DiffHunk{
		OldStart:  int(hunk.old_start),
		OldLines:  int(hunk.old_lines),
		NewStart:  int(hunk.new_start),
		NewLines:  int(hunk.new_lines),
		Header:    C.GoStringN(&hunk.header[0], C.int(hunk.header_len)),
		DiffDelta: *newDiffDeltaFromC(delta),
	}
}

type DiffLine struct {
	Origin    DiffLineType
	OldLineno int
	NewLineno int
	NumLines  int
	Content   string
	DiffHunk
}

func newDiffLineFromC(delta *C.git_diff_delta, hunk *C.git_diff_hunk, line *C.git_diff_line) *DiffLine {
	return &DiffLine{
		Origin:    DiffLineType(line.origin),
		OldLineno: int(line.old_lineno),
		NewLineno: int(line.new_lineno),
		NumLines:  int(line.num_lines),
		Content:   C.GoStringN(line.content, C.int(line.content_len)),
		DiffHunk:  *newDiffHunkFromC(delta, hunk),
	}
}

type Diff struct {
	ptr *C.git_diff
}

func newDiffFromC(ptr *C.git_diff) *Diff {
	if ptr == nil {
		return nil
	}

	diff := &Diff{
		ptr: ptr,
	}

	runtime.SetFinalizer(diff, (*Diff).Free)
	return diff
}

func (diff *Diff) Free() error {
	if diff.ptr != nil {
		return ErrInvalid
	}
	runtime.SetFinalizer(diff, nil)
	C.git_diff_free(diff.ptr)
	return nil
}

type diffForEachFileData struct {
	Callback DiffForEachFileCallback
	Error    error
}

func (diff *Diff) ForEachFile(cb DiffForEachFileCallback) error {
	if diff.ptr != nil {
		return ErrInvalid
	}

	data := &diffForEachFileData{
		Callback: cb,
	}
	ecode := C._go_git_diff_foreach(diff.ptr, 1, 0, 0, unsafe.Pointer(&data))
	if ecode < 0 {
		return data.Error
	}
	return nil
}

//export diffForEachFileCb
func diffForEachFileCb(delta *C.git_diff_delta, progress C.float, payload unsafe.Pointer) int {
	data := *diffForEachFileData(payload)

	err := data.Callback(newDiffDeltaFromC(delta))
	if err != nil {
		data.Error = err
		return -1
	}

	return 0
}

type diffForEachHunkData struct {
	Callback DiffForEachHunkCallback
	Error    error
}

type DiffForEachHunkCallback func(*DiffHunk) error

func (diff *Diff) ForEachHunk(cb DiffForEachHunkCallback) error {
	if diff.ptr != nil {
		return ErrInvalid
	}
	data := &diffForEachHunkData{
		Callback: cb,
	}
	ecode := C._go_git_diff_foreach(diff.ptr, 0, 1, 0, unsafe.Pointer(data))
	if ecode < 0 {
		return data.Error
	}
	return nil
}

//export diffForEachHunkCb
func diffForEachHunkCb(delta *C.git_diff_delta, hunk *C.git_diff_hunk, payload unsafe.Pointer) int {
	data := *diffForEachHunkData(payload)

	err := data.Callback(newDiffHunkFromC(delta, hunk))
	if err < 0 {
		data.Error = err
		return -1
	}

	return 0
}

type diffForEachLineData struct {
	Callback DiffForEachLineCallback
	Error    error
}

type DiffForEachLineCallback func(*DiffLine) error

func (diff *Diff) ForEachLine(cb DiffForEachLineCallback) error {
	if diff.ptr != nil {
		return ErrInvalid
	}

	data := &diffForEachLineData{
		Callback: cb,
	}

	ecode := C._go_git_diff_foreach(diff.ptr, 0, 0, 1, unsafe.Pointer(data))
	if ecode < 0 {
		return data.Error
	}
	return nil
}

//export diffForEachLineCb
func diffForEachLineCb(delta *C.git_diff_delta, hunk *C.git_diff_hunk, line *C.git_diff_line, payload unsafe.Pointer) int {

	data := *diffForEachLineData(payload)

	err := data.Callback(newDiffLineFromC(delta, hunk, line))
	if err != nil {
		data.Error = err
		return -1
	}

	return 0
}

func (diff *Diff) NumDeltas() (int, error) {
	if diff.ptr != nil {
		return -1, ErrInvalid
	}
	return int(C.git_diff_num_deltas(diff.ptr)), nil
}

func (diff *Diff) GetDelta(index int) (*DiffDelta, error) {
	if diff.ptr != nil {
		return nil, ErrInvalid
	}
	ptr := C.git_diff_get_delta(diff.ptr, C.size_t(index))
	if ptr == nil {
		return nil
	}

	return newDiffDeltaFromC(ptr), nil
}

func (diff *Diff) Patch(deltaIndex int) (*Patch, error) {
	if diff.ptr != nil {
		return nil, ErrInvalid
	}
	var patchPtr *C.git_patch

	ecode := C.git_patch_from_diff(&patchPtr, diff.ptr, C.size_t(deltaIndex))
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newPatchFromC(patchPtr), nil
}

func (v *Repository) DiffTreeToTree(oldTree, newTree *Tree) *Diff {
	var diffPtr *C.git_diff
	var oldPtr, newPtr *C.git_tree

	if oldTree != nil {
		oldPtr = oldTree.gitObject.ptr
	}

	if newTree != nil {
		newPtr = newTree.gitObject.ptr
	}

	C.git_diff_tree_to_tree(&diffPtr, v.ptr, oldPtr, newPtr, nil)

	return newDiff(diffPtr)
}
