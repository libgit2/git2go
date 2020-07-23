#!/bin/sh

# Since CMake cannot build the static and dynamic libraries in the same
# directory, this script helps build both static and dynamic versions of it and
# have the common flags in one place instead of split between two places.

set -e

if [ "$#" -eq "0" ]; then
	echo "Usage: $0 <--dynamic|--static>">&2
	exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && echo "${PWD}")"
VENDORED_PATH="${ROOT}/vendor/libgit2"

case "$1" in
	--static)
		BUILD_PATH="${ROOT}/static-build"
		BUILD_SHARED_LIBS=OFF
		;;

	--dynamic)
		BUILD_PATH="${ROOT}/dynamic-build"
		BUILD_SHARED_LIBS=ON
		;;

	*)
		echo "Usage: $0 <--dynamic|--static>">&2
		exit 1
		;;
esac

mkdir -p "${BUILD_PATH}/build" "${BUILD_PATH}/install/lib"

cd "${BUILD_PATH}/build" &&
cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DBUILD_SHARED_LIBS"=${BUILD_SHARED_LIBS}" \
      -DREGEX_BACKEND=builtin \
      -DCMAKE_C_FLAGS=-fPIC \
      -DCMAKE_BUILD_TYPE="RelWithDebInfo" \
      -DCMAKE_INSTALL_PREFIX="${BUILD_PATH}/install" \
      -DCMAKE_INSTALL_LIBDIR="lib" \
      "${VENDORED_PATH}" &&

exec cmake --build . --target install
