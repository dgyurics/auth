PROJECT_NAME = auth
AUTH_CONTAINER_NAME = auth-server
SECURE_CONTAINER_NAME = secure-server

.DEFAULT_GOAL := run

run:
	docker compose -p ${PROJECT_NAME} up -d

build:
	docker compose build

test:
	go test -v -race github.com/dgyurics/auth/auth-server/...

test-all: test
	go test -v -race github.com/dgyurics/auth/secure-server/...

coverage:
	go test -v -race -coverprofile=coverage.out github.com/dgyurics/auth/auth-server/...
#	go test -v -race -coverprofile=coverage.out.tmp github.com/dgyurics/auth/secure-server/...
#	cat coverage.out.tmp >> coverage.out
#	rm -f coverage.out.tmp
	go tool cover -html=coverage.out -o coverage.html

lint:
	cd auth-server && golangci-lint run ./...

# check for vulnerabilities in dependencies
# requires go install golang.org/x/vuln/cmd/govulncheck@latest
vulnerabilities:
	govulncheck ./$(AUTH_CONTAINER_NAME)/...

# start single docker container
run-container:
	docker compose up -d --no-deps $(container) --name $(container)

# rebuild single docker container
rebuild-container:
	docker compose build --no-cache $(container)

# stop docker containers which match the given name
stop-containers:
	docker kill $$(docker ps -q -f "name=$(container)")

# follow service logs in real time
logs:
	docker compose logs -f $(service)
