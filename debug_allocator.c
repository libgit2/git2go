#include "_cgo_export.h"

#include <execinfo.h>
#include <fcntl.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>

#include <git2.h>
#include <git2/common.h>
#include <git2/sys/alloc.h>

static git_allocator _go_git_system_allocator;
static git_allocator _go_git_debug_allocator;

static int __alloc_fd = -1;

static void log_alloc_event(char type, const void* ptr, size_t len, const char *file, int line) {
	void *btaddr[16];
	char buffer[8192];
	char **strings = NULL;
	size_t ptr_size = sizeof(buffer), buffer_size = 0;
	int written;

	if (type == 'D') {
		written = snprintf(buffer + buffer_size, ptr_size, "%c\t%p\n", type, ptr);
		if (written < 0 || written >= ptr_size) {
			perror("snprintf");
			abort();
		}
		ptr_size -= written;
		buffer_size += written;
	} else {
		size_t i;
		int btlen = backtrace(btaddr, 16);
		strings = backtrace_symbols(btaddr, btlen);

		written = snprintf(buffer + buffer_size, ptr_size, "%c\t%p\t%zu\t%s:%d", type, ptr, len, file, line);
		if (written < 0 || written >= ptr_size) {
			perror("snprintf");
			abort();
		}
		ptr_size -= written;
		buffer_size += written;

		for (i = 0; i < btlen; ++i) {
			written = snprintf(buffer + buffer_size, ptr_size, "\t%s", strings[i]);
			if (written < 0 || written >= ptr_size) {
				perror("snprintf");
				abort();
			}
			ptr_size -= written;
			buffer_size += written;
		}
		free(strings);

		written = snprintf(buffer + buffer_size, ptr_size, "\n");
		if (written < 0 || written >= ptr_size) {
			perror("snprintf");
			abort();
		}
		ptr_size -= written;
		buffer_size += written;
	}

	if (write(__alloc_fd, buffer, buffer_size) != buffer_size) {
		perror("write");
		abort();
	}
}

static void *_go_git_debug_allocator__malloc(size_t len, const char *file, int line)
{
	void *ptr = _go_git_system_allocator.gmalloc(len, file, line);
	if (ptr)
		log_alloc_event('A', ptr, len, file, line);
	return ptr;
}

static void *_go_git_debug_allocator__calloc(size_t nelem, size_t elsize, const char *file, int line)
{
	void *ptr = _go_git_system_allocator.gcalloc(nelem, elsize, file, line);
	if (ptr)
		log_alloc_event('A', ptr, nelem * elsize, file, line);
	return ptr;
}

static char *_go_git_debug_allocator__strdup(const char *str, const char *file, int line)
{
	char *ptr = _go_git_system_allocator.gstrdup(str, file, line);
	if (ptr)
		log_alloc_event('A', ptr, strlen(ptr) + 1, file, line);
	return ptr;
}

static char *_go_git_debug_allocator__strndup(const char *str, size_t n, const char *file, int line)
{
	char *ptr = _go_git_system_allocator.gstrndup(str, n, file, line);
	if (ptr)
		log_alloc_event('A', ptr, strlen(ptr) + 1, file, line);
	return ptr;
}

static char *_go_git_debug_allocator__substrdup(const char *start, size_t n, const char *file, int line)
{
	char *ptr = _go_git_system_allocator.gsubstrdup(start, n, file, line);
	if (ptr)
		log_alloc_event('A', ptr, strlen(ptr) + 1, file, line);
	return ptr;
}

static void *_go_git_debug_allocator__realloc(void *ptr, size_t size, const char *file, int line)
{
	void *new_ptr = _go_git_system_allocator.grealloc(ptr, size, file, line);
	if (new_ptr != ptr) {
		if (ptr)
			log_alloc_event('D', ptr, 0, NULL, 0);
		if (new_ptr)
			log_alloc_event('A', new_ptr, size, file, line);
	} else if (new_ptr) {
		log_alloc_event('R', new_ptr, size, file, line);
	}
	return new_ptr;
}

static void *_go_git_debug_allocator__reallocarray(void *ptr, size_t nelem, size_t elsize, const char *file, int line)
{
	void *new_ptr = _go_git_system_allocator.greallocarray(ptr, nelem, elsize, file, line);
	if (new_ptr != ptr) {
		if (ptr)
			log_alloc_event('D', ptr, 0, NULL, 0);
		if (new_ptr)
			log_alloc_event('A', new_ptr, nelem * elsize, file, line);
	} else if (new_ptr) {
		log_alloc_event('R', new_ptr, nelem * elsize, file, line);
	}
	return new_ptr;
}

static void *_go_git_debug_allocator__mallocarray(size_t nelem, size_t elsize, const char *file, int line)
{
	void *ptr = _go_git_system_allocator.gmallocarray(nelem, elsize, file, line);
	if (ptr)
		log_alloc_event('A', ptr, nelem * elsize, file, line);
	return ptr;
}

static void _go_git_debug_allocator__free(void *ptr)
{
	_go_git_system_allocator.gfree(ptr);
	if (ptr)
		log_alloc_event('D', ptr, 0, NULL, 0);
}

int _go_git_setup_debug_allocator(const char *log_path)
{
#if defined(LIBGIT2_STATIC)
	int error;

	__alloc_fd = open(log_path, O_CREAT | O_TRUNC | O_CLOEXEC | O_WRONLY, 0644);
	if (__alloc_fd == -1) {
		perror("open");
		return -1;
	}

	error = git_stdalloc_init_allocator(&_go_git_system_allocator);
	if (error < 0)
		return error;
	_go_git_debug_allocator.gmalloc = _go_git_debug_allocator__malloc;
	_go_git_debug_allocator.gcalloc = _go_git_debug_allocator__calloc;
	_go_git_debug_allocator.gstrdup = _go_git_debug_allocator__strdup;
	_go_git_debug_allocator.gstrndup = _go_git_debug_allocator__strndup;
	_go_git_debug_allocator.gsubstrdup = _go_git_debug_allocator__substrdup;
	_go_git_debug_allocator.grealloc = _go_git_debug_allocator__realloc;
	_go_git_debug_allocator.greallocarray = _go_git_debug_allocator__reallocarray;
	_go_git_debug_allocator.gmallocarray = _go_git_debug_allocator__mallocarray;
	_go_git_debug_allocator.gfree = _go_git_debug_allocator__free;
	error = git_libgit2_opts(GIT_OPT_SET_ALLOCATOR, &_go_git_debug_allocator);
	if (error < 0)
		return error;

	return 0;
#elif defined(LIBGIT2_DYNAMIC)
	fprintf(stderr, "debug allocator is only enabled in static builds\n");
	return -1;
#else
#error no LIBGIT2_STATIC or LIBGIT2_DYNAMIC defined!
#endif
}
