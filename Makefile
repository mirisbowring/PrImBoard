all: bin/example
test: lint unit-test

PLATFORM=local

.PHONY: build up down exec restart

config:
	@COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 COMPOSE_TARGET=config docker-compose -f docker-compose.yml config $(c)

build:
	@COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 COMPOSE_TARGET=run docker-compose -f docker-compose.yml build $(c)
	# @docker build . --target bin \
	# --output bin/ \
	# --platform ${COMPOSE_PLATFORM}

up:
	@make build
	@COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 COMPOSE_TARGET=run docker-compose -f docker-compose.yml up -d $(c)

down:
	@COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 COMPOSE_TARGET=run docker-compose -f docker-compose.yml down $(c)

stop:
	@COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 COMPOSE_TARGET=run docker-compose -f docker-compose.yml stop $(c)

exec:
	@COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 COMPOSE_TARGET=run docker-compose -f docker-compose.yml exec $(c) sh

restart:
	@make down $(c)
	@make up $(c)

rerun:
	@make stop $(c)
	@make up $(c)