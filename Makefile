default: test

build-libgit2:
	./script/build-libgit2-static.sh

test: build-libgit2
	./script/with-static.sh go test ./...

install: build-libgit2
	./script/with-static.sh go install ./...
