#!/bin/sh

set -ex

# Make sure we have the latest libgit2
if [ -d libgit2 ]; then
    cd libgit2
    git fetch origin development
    git checkout FETCH_HEAD
    cd ..
else
    git clone --depth 1 --single-branch git://github.com/libgit2/libgit2 libgit2
fi

cd libgit2
cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DBUILD_SHARED_LIBS=OFF \
      -DCMAKE_INSTALL_PREFIX=$PWD/install \
      .

make install
