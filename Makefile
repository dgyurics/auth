# build project
build:
	go build -v ./...

# run project in development mode
run:
	go run src/main.go

# test project
test:
	go test -v -race ./...

# lint project
# requires installation of linter: https://github.com/golangci/golangci-lint
lint:
	golangci-lint run

# remove go package
remove-package:
	go get $(package)@none
	go clean -cache -modcache

# start single docker container
run-container:
	docker-compose up --no-deps $(container)

# start multiple docker containers of the same image
# container_name & host port must be omitted from docker-compose.yml
run-container-scale:
	docker-compose up --no-deps --scale $(container)=$(instances)

rebuild-container:
	docker-compose build --no-cache $(container)
