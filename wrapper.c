#include "_cgo_export.h"
#include "git2.h"
#include "git2/submodule.h"
#include "git2/pack.h"

typedef int (*gogit_submodule_cbk)(git_submodule *sm, const char *name, void *payload);

int _go_git_visit_submodule(git_repository *repo, void *fct)
{
	  return git_submodule_foreach(repo, (gogit_submodule_cbk)&SubmoduleVisitor, fct);
}

int _go_git_treewalk(git_tree *tree, git_treewalk_mode mode, void *ptr)
{
	return git_tree_walk(tree, mode, (git_treewalk_cb)&CallbackGitTreeWalk, ptr);
}

int _go_git_packbuilder_foreach(git_packbuilder *pb, void *payload)
{
    return git_packbuilder_foreach(pb, (git_packbuilder_foreach_cb)&packbuilderForEachCb, payload);
}

/* EOF */
