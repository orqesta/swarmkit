IMAGE_NAME ?= orqesta/swarmkit:$(shell git rev-parse HEAD)
DOCKER_IMAGE_DIR = /workspace/swarmkit
BUILDX_CMD ?= docker buildx
.DEFAULT_GOAL := all

# Run builds with the source mounted directly. Dependencies are resolved by
# Go Modules inside the container; no additional sync service is required.
%::
	@echo "Running target $@ inside a container"
	@docker run --pull=never --rm -t \
		--volume "swarmkit-mod-cache:/go/pkg/mod" \
		--volume "$(ROOTDIR):$(DOCKER_IMAGE_DIR)" \
		--workdir "$(DOCKER_IMAGE_DIR)" \
		$(DOCKER_SWARMKIT_DOCKER_RUN_FLAGS) \
		"$(IMAGE_NAME)" make "$@"

.PHONY: shell
shell:
	@docker run --pull=never --rm -it \
		--volume "swarmkit-mod-cache:/go/pkg/mod" \
		--volume "$(ROOTDIR):$(DOCKER_IMAGE_DIR)" \
		--workdir "$(DOCKER_IMAGE_DIR)" \
		$(DOCKER_SWARMKIT_DOCKER_RUN_FLAGS) \
		"$(IMAGE_NAME)" bash

.PHONY: image
image:
	IMAGE_NAME="$(IMAGE_NAME)" $(BUILDX_CMD) bake dev

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
