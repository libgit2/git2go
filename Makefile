default: test

build-libgit2:
	./script/build-libgit2-static.sh

test-static: build-libgit2
	go run script/check-MakeGitError-thread-lock.go
	go test -tags static ./...

build-static: build-libgit2
	go build -tags static ./...

install-static: build-libgit2
	go install -tags static ./...

test:
	go run script/check-MakeGitError-thread-lock.go
	go test ./...

install:
	go install ./...
