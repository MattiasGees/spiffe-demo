.DEFAULT_GOAL := build

IMAGE_SPIFFE_RETRIEVER ?= mattiasgees/spiffe-retriever:latest
IMAGE_SERVER ?= mattiasgees/spiffe-server:latest
IMAGE_CLIENT ?= mattiasgees/spiffe-client:latest

export DOCKER_CLI_EXPERIMENTAL=enabled

.PHONY: build # Build the container image
build:
	@docker buildx create --use --name=crossplat --node=crossplat && \
	docker buildx build \
		--output "type=docker,push=false" \
		--build-arg APPPATH=spiffe-retriever \
		--tag $(IMAGE_SPIFFE_RETRIEVER) \
		--tag Dockerfile \
		.
	@docker buildx create --use --name=crossplat --node=crossplat && \
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg APPPATH=server \
		--tag $(IMAGE_SERVER) \
		--file Dockerfile \
		.
	@docker buildx create --use --name=crossplat --node=crossplat && \
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg APPPATH=client \
		--tag $(IMAGE_CLIENT) \
		--file Dockerfile \
		.

.PHONY: publish # Push all the image to the remote registry
publish:
	@docker buildx create --use --name=crossplat --node=crossplat && \
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg APPPATH=spiffe-retriever \
		--output "type=image,push=true" \
		--tag $(IMAGE_SPIFFE_RETRIEVER) \
		--file Dockerfile \
		.
	@docker buildx create --use --name=crossplat --node=crossplat && \
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg APPPATH=server \
		--output "type=image,push=true" \
		--tag $(IMAGE_SERVER) \
		--file Dockerfile \
		.
	@docker buildx create --use --name=crossplat --node=crossplat && \
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg APPPATH=client \
		--output "type=image,push=true" \
		--tag $(IMAGE_CLIENT) \
		--file Dockerfile \
		.
