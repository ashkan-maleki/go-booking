run:
	go run cmd/*.go

compose-up:
	docker compose -f docker/monitor/docker-compose.yaml \
	-f docker/monitor/docker-compose.override.yaml \
	-f docker/infra/docker-compose.yaml \
	-f docker/infra/docker-compose.override.yaml up \
	-d

compose-down:
	docker compose -f docker/monitor/docker-compose.yaml \
	-f docker/monitor/docker-compose.override.yaml \
	-f docker/infra/docker-compose.yaml \
	-f docker/infra/docker-compose.override.yaml down