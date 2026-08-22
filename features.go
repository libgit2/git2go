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
	FeatureHTTPS Feature = C.GIT_FEATURE_HTTPS

	// libgit2 was build with SSH support built-in
	FeatureSSH Feature = C.GIT_FEATURE_SSH

	// libgit2 was built with nanosecond support for files
	FeatureNSec Feature = C.GIT_FEATURE_NSEC

	// HTTP parsing; always available
	FeatureHTTPParser Feature = C.GIT_FEATURE_HTTP_PARSER

	// Regular expression support; always available
	FeatureRegex Feature = C.GIT_FEATURE_REGEX

	// Internationalization support for filename translation
	FeatureI18N Feature = C.GIT_FEATURE_I18N

	// NTLM support over HTTPS
	FeatureAuthNTLM Feature = C.GIT_FEATURE_AUTH_NTLM

	// Kerberos (SPNEGO) authentication support over HTTPS
	FeatureAuthNegotiate Feature = C.GIT_FEATURE_AUTH_NEGOTIATE

	// zlib support; always available
	FeatureCompression Feature = C.GIT_FEATURE_COMPRESSION

	// SHA1 object support; always available
	FeatureSHA1 Feature = C.GIT_FEATURE_SHA1

	// SHA256 object support
	FeatureSHA256 Feature = C.GIT_FEATURE_SHA256
)

// Features returns a bit-flag of Feature values indicating which features the
// loaded libgit2 library has.
func Features() Feature {
	features := C.git_libgit2_features()

	return Feature(features)
}
