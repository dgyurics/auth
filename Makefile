# Use this for running the project in development mode
run:
	go run src/main.go

# Remove go package
remove-package:
	go get $(package)@none
	go clean -cache -modcache

# Start single docker container
run-container:
	docker-compose up -d $(container)
