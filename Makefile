.DEFAULT_GOAL := build

IMAGE_SPIFFE_RETRIEVER ?= mattiasgees/spiffe-retriever:latest

export DOCKER_CLI_EXPERIMENTAL=enabled

.PHONY: build # Build the container image
build:
	@docker buildx create --use --name=crossplat --node=crossplat && \
	docker buildx build \
		--output "type=docker,push=false" \
		--build-arg BINARYNAME=spiffe-retriever \
		--tag $(IMAGE_SPIFFE_RETRIEVER) \
		--tag Dockerfile \
		.

.PHONY: publish # Push all the image to the remote registry
publish:
	@docker buildx create --use --name=crossplat --node=crossplat && \
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg BINARYNAME=spiffe-retriever \
		--output "type=image,push=true" \
		--tag $(IMAGE_SPIFFE_RETRIEVER) \
		--file Dockerfile \
		.
