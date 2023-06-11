run:
	go run cmd/*.go

compose-up:
	docker compose -f docker/infra/docker-compose.yaml \
	-f docker/infra/docker-compose.override.yaml \
	-f docker/tooling/docker-compose.yaml \
	-f docker/tooling/docker-compose.override.yaml up \
	-d

compose-down:
	docker compose -f docker/infra/docker-compose.yaml \
	-f docker/infra/docker-compose.override.yaml \
	-f docker/tooling/docker-compose.yaml \
	-f docker/tooling/docker-compose.override.yaml down