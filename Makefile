# run project in development mode
run:
	go run src/main.go

# test project
test:
	go test ./...

# remove go package
remove-package:
	go get $(package)@none
	go clean -cache -modcache

# start single docker container
run-container:
	docker-compose up -d $(container)

# start multiple docker containers of the same image
# note container_name and host port need to be omitted from docker-compose.yml
run-container-scale:
	docker-compose up -d --scale $(container)=$(instances)
