#!/bin/sh

set -ex

git clone --depth 1 --single-branch git://github.com/libgit2/libgit2 libgit2

cd libgit2
cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DCMAKE_INSTALL_PREFIX=$PWD/install \
      .

make install

# Let the Go build system know where to find libgit2
export LD_LIBRARY_PATH="$TMPDIR/libgit2/install/lib"
export PKG_CONFIG_PATH="$TMPDIR/libgit2/install/lib/pkgconfig"
