#!/bin/sh

#
# Install libgit2 to git2go in dynamic mode on Travis
#

set -ex

# We don't want to build libgit2 on the next branch, as we carry a
# submodule with the exact version we support
if [ "x$TRAVIS_BRANCH" = "xnext" ]; then
    exit 0
fi

cd "${HOME}"
LG2VER="0.24.0"
wget -O libgit2-${LG2VER}.tar.gz https://github.com/libgit2/libgit2/archive/v${LG2VER}.tar.gz
tar -xzvf libgit2-${LG2VER}.tar.gz
cd libgit2-${LG2VER} && mkdir build && cd build
cmake -DTHREADSAFE=ON -DBUILD_CLAR=OFF -DCMAKE_BUILD_TYPE="RelWithDebInfo" .. && make && sudo make install
sudo ldconfig
cd "${TRAVIS_BUILD_DIR}"
