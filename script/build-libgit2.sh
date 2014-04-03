#!/bin/sh

set -ex

git clone --depth 1 --single-branch git://github.com/libgit2/libgit2 libgit2

cd libgit2
cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DCMAKE_INSTALL_PREFIX=$PWD/install \
      .

make install
