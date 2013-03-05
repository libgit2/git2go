package git

/*
#include <git2.h>
#include <git2/errors.h>

extern int _go_git_treewalk(git_tree *tree, git_treewalk_mode mode, void *ptr);
*/
import "C"

import (
)

// Commit
type Commit struct {
	ptr *C.git_commit
}

func (c *Commit) Id() *Oid {
	return newOidFromC(C.git_commit_id(c.ptr))
}

func (c *Commit) Message() string {
	return C.GoString(C.git_commit_message(c.ptr))
}

func (c *Commit) Tree() (*Tree, error) {
	tree := new(Tree)

	err := C.git_commit_tree(&tree.ptr, c.ptr)
	if err < 0 {
		return nil, LastError()
	}
	return tree, nil
}

func (c *Commit) TreeId() *Oid {
	return newOidFromC(C.git_commit_tree_id(c.ptr))
}

/* TODO */
/*
func (c *Commit) Author() *Signature {
	ptr := C.git_commit_author(c.ptr)
	return &Signature{ptr}
}

func (c *Commit) Committer() *Signature {
	ptr := C.git_commit_committer(c.ptr)
	return &Signature{ptr}
}
*/
