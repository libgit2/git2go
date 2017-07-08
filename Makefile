default: test

test: build-libgit2
	go run script/check-MakeGitError-thread-lock.go
	go test ./...

install: build-libgit2
	go install ./...

build-libgit2:
	./script/build-libgit2-static.sh

install-static: build-libgit2
	go install --tags "static" ./...

test-static: build-libgit2
	go run script/check-MakeGitError-thread-lock.go
	go test --tags "static" ./...
