package git

import (
	"bytes"
	"errors"
	"runtime"
	"strconv"
)

// An equivalent of the Tree struct except this one is parsed in Go. There is no
// corresponding libgit2 object so you cannot do everything that you can with a
// Tree.
type ManagedTree struct {
	m map[string]*TreeEntry
	l []*TreeEntry
}

func (t *ManagedTree) EntryById(id *Oid) *TreeEntry {
	for _, entry := range t.l {
		if entry.Id.Equal(id) {
			return entry
		}
	}

	return nil
}

func (t *ManagedTree) EntryByName(filename string) *TreeEntry {
	return t.m[filename]
}

func (t *ManagedTree) EntryByIndex(index uint64) *TreeEntry {
	return t.l[index]
}

func (t *ManagedTree) EntryCount() uint64 {
	return uint64(len(t.l))
}

// NewManagedTree retrieves the tree with the given id and parses it.
func NewManagedTree(r *Repository, id *Oid) (*ManagedTree, error) {
	odb, err := r.Odb()
	if err != nil {
		return nil, err
	}

	obj, err := odb.Read(id)
	if err != nil {
		return nil, err
	}

	if obj.Type() != ObjectTree {
		return nil, errors.New("object is not a tree")
	}

	data := obj.Data()
	l := make([]*TreeEntry, 0, 8)
	m := make(map[string]*TreeEntry)

	var done bool
	for !done {
		spAt := bytes.IndexByte(data, ' ')
		if spAt < 0 {
			return nil, errors.New("failed to find SP after mode")
		}
		mode, err := strconv.ParseInt(string(data[:spAt]), 8, 32)
		if err != nil {
			return nil, err
		}

		data = data[spAt+1:]
		nulAt := bytes.IndexByte(data, 0)
		if nulAt < 0 {
			return nil, errors.New("failed to find NUL after filename")
		}

		var name bytes.Buffer
		name.Grow(nulAt)
		name.Write(data[:nulAt])

		data = data[nulAt+1:]
		oid := NewOidFromBytes(data)
		if len(data) > 20 {
			data = data[20:]
		} else {
			done = true
		}

		entry := &TreeEntry{
			Name:     name.String(),
			Id:       oid,
			Type:     typeFromMode(mode),
			Filemode: Filemode(mode),
		}

		l = append(l, entry)
		m[entry.Name] = entry
	}

	// This avoids the runtime from garbage-collecting 'obj' and freeing the
	// memory we're borrowing from libgit2.
	runtime.KeepAlive(obj)

	return &ManagedTree{
		l: l,
		m: m,
	}, nil
}

func typeFromMode(mode int64) ObjectType {
	switch Filemode(mode) {
	case FilemodeTree:
		return ObjectTree
	case FilemodeCommit:
		return ObjectCommit
	default:
		return ObjectBlob
	}
}
