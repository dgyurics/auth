# start application
run:
	docker-compose up -d

# start single docker container
run-container:
	docker-compose up -d --no-deps $(container) --name $(container)

# rebuild docker container
rebuild-container:
	docker-compose build --no-cache $(container)
