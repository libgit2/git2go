default: test

build-libgit2:
	./script/build-libgit2-static.sh

update-libgit2:
	cd vendor/libgit2 && \
	git fetch origin development && \
	git checkout -qf FETCH_HEAD

test: build-libgit2
	./script/with-static.sh go test ./...

install: build-libgit2
	./script/with-static.sh go install ./...
