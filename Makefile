.DEFAULT_GOAL := build

IMAGE_NAME ?= mattiasgees/spiffe-demo:latest
INIT_IMAGE_NAME ?= mattiasgees/spiffe-demo-init:latest
POSTGRES_IMAGE_NAME ?= mattiasgees/spiffe-postgres:latest
SPFFE_HELPER_IMAGE_NAME ?= mattiasgees/spiffe-helper:latest
SPIFFE_GCP_PROXY_IMAGE_NAME ?= mattiasgees/spiffe-gcp-proxy:latest
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
	docker buildx build \
		--output "type=docker,push=false" \
		--tag $(SPFFE_HELPER_IMAGE_NAME) \
		./deploy/spiffe-helper
	docker buildx build \
		--output "type=docker,push=false" \
		--tag $(SPFFE_GCP_PROXY_IMAGE_NAME) \
		./deploy/spiffe-gcp-proxy

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
	docker buildx build \
		--platform linux/amd64 \
		--output "type=image,push=true" \
		--tag $(SPFFE_HELPER_IMAGE_NAME) \
		./deploy/spiffe-helper
	docker buildx build \
		--output "type=docker,push=true" \
		--tag $(SPFFE_GCP_PROXY_IMAGE_NAME) \
		./deploy/spiffe-gcp-proxy

gcp-proxy:
	@docker buildx create --use --name=crossplat --node=crossplat && \
	docker buildx build \
		--output "type=docker,push=false" \
		--tag $(SPFFE_GCP_PROXY_IMAGE_NAME) \
		./deploy/spiffe-gcp-proxy