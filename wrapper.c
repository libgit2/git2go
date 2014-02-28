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

int _go_git_odb_foreach(git_odb *db, void *payload)
{
    return git_odb_foreach(db, (git_odb_foreach_cb)&odbForEachCb, payload);
}

git_merge_head** _go_git_make_merge_head_array(size_t len)
{
	return (git_merge_head**)malloc(sizeof(git_merge_head*) * len);
}

void _go_git_merge_head_array_set(git_merge_head** array, git_merge_head* ptr, size_t n)
{
	array[n] = ptr;
}

git_merge_head* _go_git_merge_head_array_get(git_merge_head** array, size_t n)
{
	return array[n];	
}
/* EOF */
