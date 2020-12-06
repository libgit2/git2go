#!/bin/sh

# Since CMake cannot build the static and dynamic libraries in the same
# directory, this script helps build both static and dynamic versions of it and
# have the common flags in one place instead of split between two places.

set -e

usage() {
	echo "Usage: $0 <--dynamic|--static> [--system]">&2
	exit 1
}

if [ "$#" -eq "0" ]; then
	usage
fi

ROOT=${ROOT-"$(cd "$(dirname "$0")/.." && echo "${PWD}")"}
VENDORED_PATH=${VENDORED_PATH-"${ROOT}/vendor/libgit2"}
BUILD_SYSTEM=OFF

while [ $# -gt 0 ]; do
	case "$1" in
		--static)
			BUILD_PATH="${ROOT}/static-build"
			BUILD_SHARED_LIBS=OFF
			;;

		--dynamic)
			BUILD_PATH="${ROOT}/dynamic-build"
			BUILD_SHARED_LIBS=ON
			;;

		--system)
			BUILD_SYSTEM=ON
			;;

		*)
			usage
			;;
	esac
	shift
done

if [ -z "${BUILD_SHARED_LIBS}" ]; then
	usage
fi

if [ -n "${BUILD_LIBGIT_REF}" ]; then
	git -C "${VENDORED_PATH}" checkout "${BUILD_LIBGIT_REF}"
	trap "git submodule update --init" EXIT
fi

if [ "${BUILD_SYSTEM}" = "ON" ]; then
	BUILD_INSTALL_PREFIX=${SYSTEM_INSTALL_PREFIX-"/usr"}
else
	BUILD_INSTALL_PREFIX="${BUILD_PATH}/install"
	mkdir -p "${BUILD_PATH}/install/lib"
fi

mkdir -p "${BUILD_PATH}/build" &&
cd "${BUILD_PATH}/build" &&
cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DBUILD_SHARED_LIBS"=${BUILD_SHARED_LIBS}" \
      -DREGEX_BACKEND=builtin \
      -DCMAKE_C_FLAGS=-fPIC \
      -DCMAKE_BUILD_TYPE="RelWithDebInfo" \
      -DCMAKE_INSTALL_PREFIX="${BUILD_INSTALL_PREFIX}" \
      -DCMAKE_INSTALL_LIBDIR="lib" \
      "${VENDORED_PATH}"

if which make nproc >/dev/null && [ -f Makefile ]; then
	# Make the build parallel if make is available and cmake used Makefiles.
	exec make "-j$(nproc --all)" install
else
	exec cmake --build . --target install
fi
