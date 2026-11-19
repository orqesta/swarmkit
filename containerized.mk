IMAGE_NAME ?= docker/swarmkit
GOPATH ?= /go
DOCKER_IMAGE_DIR = $(GOPATH)/src/github.com/docker/swarmkit
BUILDX_CMD ?= docker buildx

# Run builds with the source mounted directly. Dependencies are resolved by
# Go Modules inside the container; no additional sync service is required.
%::
	@echo "Running target $@ inside a container"
	@$(MAKE) ensure_image_exists
	@docker run --rm -t \
		--volume "swarmkit-cache:$(GOPATH)" \
		--volume "$(ROOTDIR):$(DOCKER_IMAGE_DIR)" \
		$(DOCKER_SWARMKIT_DOCKER_RUN_FLAGS) \
		"$(IMAGE_NAME)" make "$@"

.PHONY: shell
shell: ensure_image_exists
	@docker run --rm -it \
		--volume "swarmkit-cache:$(GOPATH)" \
		--volume "$(ROOTDIR):$(DOCKER_IMAGE_DIR)" \
		$(DOCKER_SWARMKIT_DOCKER_RUN_FLAGS) \
		"$(IMAGE_NAME)" bash

.PHONY: image
image:
	docker build -t "$(IMAGE_NAME)" .

.PHONY: ensure_image_exists
ensure_image_exists:
	@docker image inspect "$(IMAGE_NAME)" >/dev/null 2>&1 || $(MAKE) image

.PHONY: fmt-proto
fmt-proto:
	$(BUILDX_CMD) bake fmt-proto

.PHONY: lint
lint:
	$(BUILDX_CMD) bake lint

.PHONY: module-validate
module-validate:
	$(BUILDX_CMD) bake module-validate

.PHONY: generate-validate
generate-validate:
	$(BUILDX_CMD) bake generate-validate
