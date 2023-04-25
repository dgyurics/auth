# start single docker container
run-container:
	docker-compose up -d --no-deps $(container)

# start multiple docker containers of the same image
# container_name & host port must be omitted from docker-compose.yml
run-container-scale:
	docker-compose up --no-deps --scale $(container)=$(instances)

rebuild-container:
	docker-compose build --no-cache $(container)
