package git

/*
#include <git2.h>
*/
import "C"

type Feature int

const (
	// libgit2 was built with threading support
	FeatureThreads Feature = C.GIT_FEATURE_THREADS

	// libgit2 was built with HTTPS support built-in
	FeatureHttps Feature = C.GIT_FEATURE_HTTPS

	// libgit2 was build with SSH support built-in
	FeatureSsh Feature = C.GIT_FEATURE_SSH

	// libgit2 was built with nanosecond support for files
	FeatureNSec Feature = C.GIT_FEATURE_NSEC
)

// Features returns a bit-flag of Feature values indicating which features the
// loaded libgit2 library has.
func Features() Feature {
	features := C.git_libgit2_features()

	return Feature(features)
}
