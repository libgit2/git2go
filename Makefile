default: test

build-libgit2:
	./script/build-libgit2-static.sh
	cat ./vendor/libgit2/libgit2.pc
	cat ./vendor/install/lib/pkgconfig/libgit2.pc

test: build-libgit2
	./script/with-static.sh go test ./...

install: build-libgit2
	./script/with-static.sh go install ./...
