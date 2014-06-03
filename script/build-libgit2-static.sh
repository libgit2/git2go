#!/bin/sh

set -ex

VENDORED_PATH=vendor/libgit2

# Make sure we have the latest libgit2
if [ ! -d $VENDORED_PATH ]; then
    git clone --depth 1 --single-branch git://github.com/libgit2/libgit2 $VENDORED_PATH
fi

cd $VENDORED_PATH

cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DBUILD_SHARED_LIBS=OFF \
      -DCMAKE_INSTALL_PREFIX=../install \
      .

make install
