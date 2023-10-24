PROJECT_NAME = auth

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
	govulncheck ./auth-server/...

# start single docker container
run-ui:
	docker compose up -d --no-deps ui
run-nginx:
	docker compose up -d --no-deps nginx
run-auth:
	docker compose up -d --no-deps auth
run-secure:
	docker compose up -d --no-deps secure
run-postgres:
	docker compose up -d --no-deps postgres
run-redis:
	docker compose up -d --no-deps redis

# rebuild single docker container
rebuild-ui:
	docker compose build --no-cache ui
rebuild-nginx:
	docker compose build --no-cache nginx
rebuild-auth:
	docker compose build --no-cache auth
rebuild-secure:
	docker compose build --no-cache secure
rebuild-postgres:
	docker compose build --no-cache postgres
rebuild-redis:
	docker compose build --no-cache redis

# stop docker containers which match the given name
stop-containers:
	docker kill $$(docker ps -q -f "name=$(container)")

# stop all docker containers
stop:
	docker stop $$(docker ps -q)

# follow service logs in real time
logs:
	docker compose logs -f $(service)
