default: test

build-libgit2:
	./script/build-libgit2-static.sh

test: build-libgit2
	go run script/check-MakeGitError-thread-lock.go
	go test ./...

install: build-libgit2
	go install ./...
