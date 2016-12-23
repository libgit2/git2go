default: test

test:
	go run script/check-MakeGitError-thread-lock.go
	go test ./...

install: build-libgit2
	go install ./...
