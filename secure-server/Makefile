.DEFAULT_GOAL := run

# build project
build:
	go build -v -o secureserver ./cmd/main.go

# run in development mode
run:
	go run ./cmd/main.go

# lint project
# requires installation of golangci-lint https://github.com/golangci/golangci-lint
lint:
	golangci-lint run

# remove single package
remove-package:
	go get $(package)@none
	go clean -cache -modcache
