# build project
build:
	go build -v -o authserver ./cmd/main.go

# run in development mode
run:
	go run ./cmd/main.go

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


# TODO add govulncheck https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck
