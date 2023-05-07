# start application
run:
	docker compose -p auth up -d

build:
	docker compose build

# test application
test:
	go test -v -race github.com/dgyurics/auth/auth-server/... \
	&& go test -v -race github.com/dgyurics/auth/secure-server/...

# lint application
lint:
	cd auth-server && golangci-lint run ./...

# start single docker container
run-container:
	docker compose up -d --no-deps $(container) --name $(container)

# rebuild docker container
rebuild-container:
	docker compose build --no-cache $(container)

# stop docker containers which match the given name
stop-containers:
	docker kill $$(docker ps -q -f "name=$(container)")

# follow service logs in real time
logs:
	docker compose logs -f $(service)
