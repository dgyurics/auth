# build project
build:
	go build -v ./...

# run in development mode
run:
	go run github.com/dgyurics/auth/auth-server

# run tests
test:
	go test -v -race ./...

# lint project
# requires installation of golangci-lint https://github.com/golangci/golangci-lint
lint:
	golangci-lint run

# remove single package
remove-package:
	go get $(package)@none
	go clean -cache -modcache
