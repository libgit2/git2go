package git

/*
#include <git2.h>

extern int _go_git_diff_foreach(git_diff *diff, int eachFile, int eachHunk, int eachLine, void *payload);
extern void _go_git_setup_diff_notify_callbacks(git_diff_options* opts);
*/
import "C"
import (
	"errors"
	"runtime"
	"unsafe"
)

type DiffFlag int

const (
	DiffFlagBinary    DiffFlag = C.GIT_DIFF_FLAG_BINARY
	DiffFlagNotBinary DiffFlag = C.GIT_DIFF_FLAG_NOT_BINARY
	DiffFlagValidOid  DiffFlag = C.GIT_DIFF_FLAG_VALID_ID
)

type Delta int

const (
	DeltaUnmodified Delta = C.GIT_DELTA_UNMODIFIED
	DeltaAdded      Delta = C.GIT_DELTA_ADDED
	DeltaDeleted    Delta = C.GIT_DELTA_DELETED
	DeltaModified   Delta = C.GIT_DELTA_MODIFIED
	DeltaRenamed    Delta = C.GIT_DELTA_RENAMED
	DeltaCopied     Delta = C.GIT_DELTA_COPIED
	DeltaIgnored    Delta = C.GIT_DELTA_IGNORED
	DeltaUntracked  Delta = C.GIT_DELTA_UNTRACKED
	DeltaTypeChange Delta = C.GIT_DELTA_TYPECHANGE
)

type DiffLineType int

const (
	DiffLineContext      DiffLineType = C.GIT_DIFF_LINE_CONTEXT
	DiffLineAddition     DiffLineType = C.GIT_DIFF_LINE_ADDITION
	DiffLineDeletion     DiffLineType = C.GIT_DIFF_LINE_DELETION
	DiffLineContextEOFNL DiffLineType = C.GIT_DIFF_LINE_CONTEXT_EOFNL
	DiffLineAddEOFNL     DiffLineType = C.GIT_DIFF_LINE_ADD_EOFNL
	DiffLineDelEOFNL     DiffLineType = C.GIT_DIFF_LINE_DEL_EOFNL

	DiffLineFileHdr DiffLineType = C.GIT_DIFF_LINE_FILE_HDR
	DiffLineHunkHdr DiffLineType = C.GIT_DIFF_LINE_HUNK_HDR
	DiffLineBinary  DiffLineType = C.GIT_DIFF_LINE_BINARY
)

type DiffFile struct {
	Path  string
	Oid   *Oid
	Size  int
	Flags DiffFlag
	Mode  uint16
}

func diffFileFromC(file *C.git_diff_file) DiffFile {
	return DiffFile{
		Path:  C.GoString(file.path),
		Oid:   newOidFromC(&file.id),
		Size:  int(file.size),
		Flags: DiffFlag(file.flags),
		Mode:  uint16(file.mode),
	}
}

type DiffDelta struct {
	Status     Delta
	Flags      DiffFlag
	Similarity uint16
	OldFile    DiffFile
	NewFile    DiffFile
}

func diffDeltaFromC(delta *C.git_diff_delta) DiffDelta {
	return DiffDelta{
		Status:     Delta(delta.status),
		Flags:      DiffFlag(delta.flags),
		Similarity: uint16(delta.similarity),
		OldFile:    diffFileFromC(&delta.old_file),
		NewFile:    diffFileFromC(&delta.new_file),
	}
}

type DiffHunk struct {
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Header   string
}

func diffHunkFromC(delta *C.git_diff_delta, hunk *C.git_diff_hunk) DiffHunk {
	return DiffHunk{
		OldStart: int(hunk.old_start),
		OldLines: int(hunk.old_lines),
		NewStart: int(hunk.new_start),
		NewLines: int(hunk.new_lines),
		Header:   C.GoStringN(&hunk.header[0], C.int(hunk.header_len)),
	}
}

type DiffLine struct {
	Origin    DiffLineType
	OldLineno int
	NewLineno int
	NumLines  int
	Content   string
}

func diffLineFromC(delta *C.git_diff_delta, hunk *C.git_diff_hunk, line *C.git_diff_line) DiffLine {
	return DiffLine{
		Origin:    DiffLineType(line.origin),
		OldLineno: int(line.old_lineno),
		NewLineno: int(line.new_lineno),
		NumLines:  int(line.num_lines),
		Content:   C.GoStringN(line.content, C.int(line.content_len)),
	}
}

type Diff struct {
	ptr *C.git_diff
}

func (diff *Diff) NumDeltas() (int, error) {
	if diff.ptr == nil {
		return -1, ErrInvalid
	}
	return int(C.git_diff_num_deltas(diff.ptr)), nil
}

func (diff *Diff) GetDelta(index int) (DiffDelta, error) {
	if diff.ptr == nil {
		return DiffDelta{}, ErrInvalid
	}
	ptr := C.git_diff_get_delta(diff.ptr, C.size_t(index))
	return diffDeltaFromC(ptr), nil
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
	if diff.ptr == nil {
		return ErrInvalid
	}
	runtime.SetFinalizer(diff, nil)
	C.git_diff_free(diff.ptr)
	diff.ptr = nil
	return nil
}

type diffForEachData struct {
	FileCallback DiffForEachFileCallback
	HunkCallback DiffForEachHunkCallback
	LineCallback DiffForEachLineCallback
	Error        error
}

type DiffForEachFileCallback func(DiffDelta, float64) (DiffForEachHunkCallback, error)

type DiffDetail int

const (
	DiffDetailFiles DiffDetail = iota
	DiffDetailHunks
	DiffDetailLines
)

func (diff *Diff) ForEach(cbFile DiffForEachFileCallback, detail DiffDetail) error {
	if diff.ptr == nil {
		return ErrInvalid
	}

	intHunks := C.int(0)
	if detail >= DiffDetailHunks {
		intHunks = C.int(1)
	}

	intLines := C.int(0)
	if detail >= DiffDetailLines {
		intLines = C.int(1)
	}

	data := &diffForEachData{
		FileCallback: cbFile,
	}
	ecode := C._go_git_diff_foreach(diff.ptr, 1, intHunks, intLines, unsafe.Pointer(data))
	if ecode < 0 {
		return data.Error
	}
	return nil
}

//export diffForEachFileCb
func diffForEachFileCb(delta *C.git_diff_delta, progress C.float, payload unsafe.Pointer) int {
	data := (*diffForEachData)(payload)

	data.HunkCallback = nil
	if data.FileCallback != nil {
		cb, err := data.FileCallback(diffDeltaFromC(delta), float64(progress))
		if err != nil {
			data.Error = err
			return -1
		}
		data.HunkCallback = cb
	}

	return 0
}

type DiffForEachHunkCallback func(DiffHunk) (DiffForEachLineCallback, error)

//export diffForEachHunkCb
func diffForEachHunkCb(delta *C.git_diff_delta, hunk *C.git_diff_hunk, payload unsafe.Pointer) int {
	data := (*diffForEachData)(payload)

	data.LineCallback = nil
	if data.HunkCallback != nil {
		cb, err := data.HunkCallback(diffHunkFromC(delta, hunk))
		if err != nil {
			data.Error = err
			return -1
		}
		data.LineCallback = cb
	}

	return 0
}

type DiffForEachLineCallback func(DiffLine) error

//export diffForEachLineCb
func diffForEachLineCb(delta *C.git_diff_delta, hunk *C.git_diff_hunk, line *C.git_diff_line, payload unsafe.Pointer) int {

	data := (*diffForEachData)(payload)

	err := data.LineCallback(diffLineFromC(delta, hunk, line))
	if err != nil {
		data.Error = err
		return -1
	}

	return 0
}

func (diff *Diff) Patch(deltaIndex int) (*Patch, error) {
	if diff.ptr == nil {
		return nil, ErrInvalid
	}
	var patchPtr *C.git_patch

	ecode := C.git_patch_from_diff(&patchPtr, diff.ptr, C.size_t(deltaIndex))
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newPatchFromC(patchPtr), nil
}

type DiffOptionsFlag int

const (
	DiffNormal                 DiffOptionsFlag = C.GIT_DIFF_NORMAL
	DiffReverse                DiffOptionsFlag = C.GIT_DIFF_REVERSE
	DiffIncludeIgnored         DiffOptionsFlag = C.GIT_DIFF_INCLUDE_IGNORED
	DiffRecurseIgnoredDirs     DiffOptionsFlag = C.GIT_DIFF_RECURSE_IGNORED_DIRS
	DiffIncludeUntracked       DiffOptionsFlag = C.GIT_DIFF_INCLUDE_UNTRACKED
	DiffRecurseUntracked       DiffOptionsFlag = C.GIT_DIFF_RECURSE_UNTRACKED_DIRS
	DiffIncludeUnmodified      DiffOptionsFlag = C.GIT_DIFF_INCLUDE_UNMODIFIED
	DiffIncludeTypeChange      DiffOptionsFlag = C.GIT_DIFF_INCLUDE_TYPECHANGE
	DiffIncludeTypeChangeTrees DiffOptionsFlag = C.GIT_DIFF_INCLUDE_TYPECHANGE_TREES
	DiffIgnoreFilemode         DiffOptionsFlag = C.GIT_DIFF_IGNORE_FILEMODE
	DiffIgnoreSubmodules       DiffOptionsFlag = C.GIT_DIFF_IGNORE_SUBMODULES
	DiffIgnoreCase             DiffOptionsFlag = C.GIT_DIFF_IGNORE_CASE

	DiffDisablePathspecMatch    DiffOptionsFlag = C.GIT_DIFF_DISABLE_PATHSPEC_MATCH
	DiffSkipBinaryCheck         DiffOptionsFlag = C.GIT_DIFF_SKIP_BINARY_CHECK
	DiffEnableFastUntrackedDirs DiffOptionsFlag = C.GIT_DIFF_ENABLE_FAST_UNTRACKED_DIRS

	DiffForceText   DiffOptionsFlag = C.GIT_DIFF_FORCE_TEXT
	DiffForceBinary DiffOptionsFlag = C.GIT_DIFF_FORCE_BINARY

	DiffIgnoreWhitespace       DiffOptionsFlag = C.GIT_DIFF_IGNORE_WHITESPACE
	DiffIgnoreWhitespaceChange DiffOptionsFlag = C.GIT_DIFF_IGNORE_WHITESPACE_CHANGE
	DiffIgnoreWitespaceEol     DiffOptionsFlag = C.GIT_DIFF_IGNORE_WHITESPACE_EOL

	DiffShowUntrackedContent DiffOptionsFlag = C.GIT_DIFF_SHOW_UNTRACKED_CONTENT
	DiffShowUnmodified       DiffOptionsFlag = C.GIT_DIFF_SHOW_UNMODIFIED
	DiffPatience             DiffOptionsFlag = C.GIT_DIFF_PATIENCE
	DiffMinimal              DiffOptionsFlag = C.GIT_DIFF_MINIMAL
)

type DiffNotifyCallback func(diffSoFar *Diff, deltaToAdd DiffDelta, matchedPathspec string) error

type DiffOptions struct {
	Flags            DiffOptionsFlag
	IgnoreSubmodules SubmoduleIgnore
	Pathspec         []string
	NotifyCallback   DiffNotifyCallback

	ContextLines   uint16
	InterhunkLines uint16
	IdAbbrev       uint16

	MaxSize int

	OldPrefix string
	NewPrefix string
}

func DefaultDiffOptions() (DiffOptions, error) {
	opts := C.git_diff_options{}
	ecode := C.git_diff_init_options(&opts, C.GIT_DIFF_OPTIONS_VERSION)
	if ecode < 0 {
		return DiffOptions{}, MakeGitError(ecode)
	}

	return DiffOptions{
		Flags:            DiffOptionsFlag(opts.flags),
		IgnoreSubmodules: SubmoduleIgnore(opts.ignore_submodules),
		Pathspec:         makeStringsFromCStrings(opts.pathspec.strings, int(opts.pathspec.count)),
		ContextLines:     uint16(opts.context_lines),
		InterhunkLines:   uint16(opts.interhunk_lines),
		IdAbbrev:         uint16(opts.id_abbrev),
		MaxSize:          int(opts.max_size),
	}, nil
}

var (
	ErrDeltaSkip = errors.New("Skip delta")
)

type diffNotifyData struct {
	Callback DiffNotifyCallback
	Diff     *Diff
	Error    error
}

//export diffNotifyCb
func diffNotifyCb(_diff_so_far unsafe.Pointer, delta_to_add *C.git_diff_delta, matched_pathspec *C.char, payload unsafe.Pointer) int {
	diff_so_far := (*C.git_diff)(_diff_so_far)
	data := (*diffNotifyData)(payload)
	if data != nil {
		if data.Diff == nil {
			data.Diff = newDiffFromC(diff_so_far)
		}

		err := data.Callback(data.Diff, diffDeltaFromC(delta_to_add), C.GoString(matched_pathspec))

		if err == ErrDeltaSkip {
			return 1
		} else if err != nil {
			data.Error = err
			return -1
		} else {
			return 0
		}
	}
	return 0
}

func (v *Repository) DiffTreeToTree(oldTree, newTree *Tree, opts *DiffOptions) (*Diff, error) {
	var diffPtr *C.git_diff
	var oldPtr, newPtr *C.git_tree

	if oldTree != nil {
		oldPtr = oldTree.cast_ptr
	}

	if newTree != nil {
		newPtr = newTree.cast_ptr
	}

	cpathspec := C.git_strarray{}
	var copts *C.git_diff_options
	var notifyData *diffNotifyData
	if opts != nil {
		notifyData = &diffNotifyData{
			Callback: opts.NotifyCallback,
		}
		if opts.Pathspec != nil {
			cpathspec.count = C.size_t(len(opts.Pathspec))
			cpathspec.strings = makeCStringsFromStrings(opts.Pathspec)
			defer freeStrarray(&cpathspec)
		}

		copts = &C.git_diff_options{
			version:           C.GIT_DIFF_OPTIONS_VERSION,
			flags:             C.uint32_t(opts.Flags),
			ignore_submodules: C.git_submodule_ignore_t(opts.IgnoreSubmodules),
			pathspec:          cpathspec,
			context_lines:     C.uint16_t(opts.ContextLines),
			interhunk_lines:   C.uint16_t(opts.InterhunkLines),
			id_abbrev:         C.uint16_t(opts.IdAbbrev),
			max_size:          C.git_off_t(opts.MaxSize),
		}

		if opts.NotifyCallback != nil {
			C._go_git_setup_diff_notify_callbacks(copts)
			copts.notify_payload = unsafe.Pointer(notifyData)
		}
	}

	ecode := C.git_diff_tree_to_tree(&diffPtr, v.ptr, oldPtr, newPtr, copts)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	if notifyData != nil && notifyData.Diff != nil {
		return notifyData.Diff, nil
	}
	return newDiffFromC(diffPtr), nil
}

func (v *Repository) DiffTreeToWorkdir(oldTree *Tree, opts *DiffOptions) (*Diff, error) {
	var diffPtr *C.git_diff
	var oldPtr *C.git_tree

	if oldTree != nil {
		oldPtr = oldTree.cast_ptr
	}

	cpathspec := C.git_strarray{}
	var copts *C.git_diff_options
	var notifyData *diffNotifyData
	if opts != nil {
		notifyData = &diffNotifyData{
			Callback: opts.NotifyCallback,
		}
		if opts.Pathspec != nil {
			cpathspec.count = C.size_t(len(opts.Pathspec))
			cpathspec.strings = makeCStringsFromStrings(opts.Pathspec)
			defer freeStrarray(&cpathspec)
		}

		copts = &C.git_diff_options{
			version:           C.GIT_DIFF_OPTIONS_VERSION,
			flags:             C.uint32_t(opts.Flags),
			ignore_submodules: C.git_submodule_ignore_t(opts.IgnoreSubmodules),
			pathspec:          cpathspec,
			context_lines:     C.uint16_t(opts.ContextLines),
			interhunk_lines:   C.uint16_t(opts.InterhunkLines),
			id_abbrev:         C.uint16_t(opts.IdAbbrev),
			max_size:          C.git_off_t(opts.MaxSize),
		}

		if opts.NotifyCallback != nil {
			C._go_git_setup_diff_notify_callbacks(copts)
			copts.notify_payload = unsafe.Pointer(notifyData)
		}
	}

	ecode := C.git_diff_tree_to_workdir(&diffPtr, v.ptr, oldPtr, copts)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	if notifyData != nil && notifyData.Diff != nil {
		return notifyData.Diff, nil
	}
	return newDiffFromC(diffPtr), nil

}
