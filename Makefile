# start application
run:
	docker-compose up -d

# test application
test:
	go test -v -race github.com/dgyurics/auth/auth-server/... \
	&& go test -v -race github.com/dgyurics/auth/secure-server/...

# lint application
lint:
	cd auth-server && golangci-lint run ./...

# start single docker container
run-container:
	docker-compose up -d --no-deps $(container) --name $(container)

# rebuild docker container
rebuild-container:
	docker-compose build --no-cache $(container)
