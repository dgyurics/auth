# Use this for running the project in development mode
run:
	go run src/main.go

# Remove go package
remove:
	go get $(package)@none
	go clean -cache -modcache
