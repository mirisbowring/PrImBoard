all: bin/example
test: lint unit-test

PLATFORM=local

.PHONY: build up

config:
	@COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 COMPOSE_TARGET=config docker-compose -f docker-compose.yml config $(c)

build:
	@COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 COMPOSE_TARGET=run docker-compose -f docker-compose.yml build $(c)
	# @docker build . --target bin \
	# --output bin/ \
	# --platform ${COMPOSE_PLATFORM}

up:
	@make build
	@COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 COMPOSE_TARGET=run docker-compose -f docker-compose.yml up $(c)

down:
	@COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 COMPOSE_TARGET=run docker-compose -f docker-compose.yml down $(c)

exec:
	@COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 COMPOSE_TARGET=run docker-compose -f docker-compose.yml exec $(c) sh