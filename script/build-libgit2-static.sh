#!/bin/sh

set -ex

ROOT="$(cd "$(dirname $0)/.." && echo "${PWD}")"
BUILD_PATH="${ROOT}/static-build"
VENDORED_PATH="${ROOT}/vendor/libgit2"

mkdir -p "${BUILD_PATH}/build" "${BUILD_PATH}/install/lib"

cd "${BUILD_PATH}/build" &&
cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DBUILD_SHARED_LIBS=OFF \
      -DCMAKE_C_FLAGS=-fPIC \
      -DCMAKE_BUILD_TYPE="RelWithDebInfo" \
      -DCMAKE_INSTALL_PREFIX="${BUILD_PATH}/install" \
      "${VENDORED_PATH}" &&

cmake --build . --target install
