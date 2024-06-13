.DEFAULT_GOAL := build

IMAGE_NAME ?= mattiasgees/spiffe-demo:latest
INIT_IMAGE_NAME ?= mattiasgees/spiffe-demo-init:latest
POSTGRES_IMAGE_NAME ?= mattiasgees/spiffe-postgres:latest

export DOCKER_CLI_EXPERIMENTAL=enabled

.PHONY: build # Build the container image
build:
	@docker buildx create --use --name=crossplat --node=crossplat && \
	docker buildx build \
		--output "type=docker,push=false" \
		--tag $(IMAGE_NAME) \
		--file Dockerfile \
		.
	docker buildx build \
		--output "type=docker,push=false" \
		--tag $(INIT_IMAGE_NAME) \
		./deploy/initcontainer
	docker buildx build \
		--output "type=docker,push=false" \
		--tag $(POSTGRES_IMAGE_NAME) \
		./deploy/postgresql

.PHONY: publish # Push all the image to the remote registry
publish:
	@docker buildx create --use --name=crossplat --node=crossplat && \
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--output "type=image,push=true" \
		--tag $(IMAGE_NAME) \
		--file Dockerfile \
		.
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--output "type=image,push=true" \
		--tag $(INIT_IMAGE_NAME) \
		./deploy/initcontainer
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--output "type=image,push=true" \
		--tag $(POSTGRES_IMAGE_NAME) \
		./deploy/postgresql
