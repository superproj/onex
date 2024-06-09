# ==============================================================================
# Makefile helper functions for docker image
#

DOCKER := docker
DOCKER_SUPPORTED_API_VERSION ?= 1.32

EXTRA_ARGS ?= --no-cache
_DOCKER_BUILD_EXTRA_ARGS :=

# Track code version with Docker Label.
DOCKER_LABELS ?= git-describe="$(shell date -u +v%Y%m%d)-$(shell git describe --tags --always --dirty)"

ifdef HTTP_PROXY
_DOCKER_BUILD_EXTRA_ARGS += --build-arg HTTP_PROXY=${HTTP_PROXY}
endif

ifneq ($(EXTRA_ARGS), )
_DOCKER_BUILD_EXTRA_ARGS += $(EXTRA_ARGS)
endif

# Determine image files by looking into cmd/*
CMD_DIRS ?= $(wildcard ${ONEX_ROOT}/cmd/*)
# Determine images names by stripping out the dir names.
# Filter out directories without Go files, as these directories cannot be compiled to build a docker image.
IMAGES ?= $(filter-out tools, $(foreach dir, $(CMD_DIRS), $(notdir $(if $(wildcard $(dir)/*.go), $(dir),))))
ifeq (${IMAGES},)
  $(error Could not determine IMAGES, set ONEX_ROOT or run in source dir)
endif

.PHONY: image.verify
image.verify: ## Verify docker version.
	$(eval API_VERSION := $(shell $(DOCKER) version | grep -E 'API version: {1,6}[0-9]' | head -n1 | awk '{print $$3} END { if (NR==0) print 0}' ))
	$(eval PASS := $(shell echo "$(API_VERSION) > $(DOCKER_SUPPORTED_API_VERSION)" | bc))
	@if [ $(PASS) -ne 1 ]; then \
		$(DOCKER) -v ;\
		echo "Unsupported docker version. Docker API version should be greater than $(DOCKER_SUPPORTED_API_VERSION)"; \
		exit 1; \
	fi

.PHONY: image.daemon.verify
image.daemon.verify: ## Verify docker daemon version.
	$(eval PASS := $(shell $(DOCKER) version | grep -q -E 'Experimental: {1,5}true' && echo 1 || echo 0))
	@if [ $(PASS) -ne 1 ]; then \
		echo "Experimental features of Docker daemon is not enabled. Please add \"experimental\": true in '/etc/docker/daemon.json' and then restart Docker daemon."; \
		exit 1; \
	fi

.PHONY: image.dockerfile
image.dockerfile: $(addprefix image.dockerfile., $(IMAGES)) ## Generate all dockerfiles.

.PHONY: image.dockerfile.%
image.dockerfile.%: ## Generate specified dockerfiles.
	$(eval IMAGE := $(lastword $(subst ., ,$*)))
	# Set a unified environment variable file
	@$(SCRIPTS_DIR)/gen-dockerfile.sh $(GENERATED_DOCKERFILE_DIR) $(IMAGE)
ifeq ($(V),1)
	echo "DBG: Generating Dockerfile at $(GENERATED_DOCKERFILE_DIR)/$(IMAGE)"
endif

.PHONY: image.build
image.build: image.verify go.build.verify $(addprefix image.build., $(addprefix $(IMAGE_PLAT)., $(IMAGES))) ## Build all docker images.

.PHONY: image.build.multiarch
image.build.multiarch: image.verify go.build.verify $(foreach p,$(PLATFORMS),$(addprefix image.build., $(addprefix $(p)., $(IMAGES)))) ## Build all docker images with all supported arch.

.PHONY: image.build.%
ifneq (${MULTISTAGE},1)
image.build.%: go.build.% image.dockerfile.% ## Build specified docker image.
	$(eval IMAGE := $(word 2,$(subst ., ,$*)))
	$(eval IMAGE_PLAT := $(subst _,/,$(PLATFORM)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	$(eval DOCKERFILE := Dockerfile)
	$(eval DST_DIR := $(TMP_DIR)/$(IMAGE))
	@echo "===========> Building docker image $(IMAGE) $(VERSION) for $(IMAGE_PLAT)"

	@mkdir -p $(TMP_DIR)/$(IMAGE)
	@cp -r $(OUTPUT_DIR)/platforms/$(IMAGE_PLAT)/$(IMAGE) $(TMP_DIR)/$(IMAGE)/
else
# TODO: onex-allinone, does not support multi-stage builds. Need to adjust it.
image.build.%: image.dockerfile.% ## Build specified docker image in multistage way.
	$(eval IMAGE := $(word 2,$(subst ., ,$*)))
	$(eval IMAGE_PLAT := $(subst _,/,$(PLATFORM)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	$(eval DOCKERFILE := Dockerfile.multistage)
	$(eval DST_DIR := $(ONEX_ROOT))
	@echo "===========> Building docker image $(IMAGE) $(VERSION) for $(IMAGE_PLAT)"
endif
	@export OUTPUT_DIR=$(OUTPUT_DIR)
	@if [ -f  $(ONEX_ROOT)/build/docker/$(IMAGE)/build.sh ] ; then \
		DST_DIR=$(DST_DIR) OUTPUT_DIR=$(OUTPUT_DIR) IMAGE_PLAT=${IMAGE_PLAT} \
		$(ONEX_ROOT)/build/docker/$(IMAGE)/build.sh ; \
	fi
	$(eval BUILD_SUFFIX := $(_DOCKER_BUILD_EXTRA_ARGS) --pull \
		-f $(GENERATED_DOCKERFILE_DIR)/$(IMAGE)/$(DOCKERFILE) \
		--build-arg OS=$(OS) \
		--build-arg ARCH=$(ARCH) \
		--build-arg goproxy=$($(GO) env GOPROXY) \
		--label $(DOCKER_LABELS) \
		-t $(REGISTRY_PREFIX)/$(IMAGE)-$(ARCH):$(VERSION) \
		$(DST_DIR))
	@if [ $(shell $(GO) env GOARCH) != $(ARCH) ] ; then \
		$(MAKE) image.daemon.verify ; \
		$(DOCKER) build --platform $(IMAGE_PLAT) $(BUILD_SUFFIX) ; \
	else \
		$(DOCKER) build $(BUILD_SUFFIX) ; \
	fi
	@-rm -rf $(TMP_DIR)/$(IMAGE)

.PHONY: image.push
image.push: image.verify go.build.verify $(addprefix image.push., $(addprefix $(IMAGE_PLAT)., $(IMAGES))) ## Build and push all docker images to docker registry.

.PHONY: image.push.multiarch
image.push.multiarch: image.verify go.build.verify $(foreach p,$(PLATFORMS),$(addprefix image.push., $(addprefix $(p)., $(IMAGES)))) ## Build and push all docker with supported arch to docker registry.

.PHONY: image.push.%
image.push.%: image.build.% ## Build and push specified docker image.
	@echo "===========> Pushing image $(IMAGE) $(VERSION) to $(REGISTRY_PREFIX)"
	$(DOCKER) push $(REGISTRY_PREFIX)/$(IMAGE)-$(ARCH):$(VERSION)

.PHONY: image.manifest.push
image.manifest.push: export DOCKER_CLI_EXPERIMENTAL := enabled
image.manifest.push: image.verify go.build.verify $(addprefix image.manifest.push., $(addprefix $(IMAGE_PLAT)., $(IMAGES))) ## Build and push docker image manifest.

.PHONY: image.manifest.push.%
image.manifest.push.%: image.push.% image.manifest.remove.%
	@echo "===========> Pushing manifest $(IMAGE) $(VERSION) to $(REGISTRY_PREFIX) and then remove the local manifest list"
	@$(DOCKER) manifest create $(REGISTRY_PREFIX)/$(IMAGE):$(VERSION) \
		$(REGISTRY_PREFIX)/$(IMAGE)-$(ARCH):$(VERSION)
	@$(DOCKER) manifest annotate $(REGISTRY_PREFIX)/$(IMAGE):$(VERSION) \
		$(REGISTRY_PREFIX)/$(IMAGE)-$(ARCH):$(VERSION) \
		--os $(OS) --arch ${ARCH}
	@$(DOCKER) manifest push --purge $(REGISTRY_PREFIX)/$(IMAGE):$(VERSION)

# Docker cli has a bug: https://github.com/docker/cli/issues/954
# If you find your manifests were not updated,
# Please manually delete them in $HOME/.docker/manifests/
# and re-run.
.PHONY: image.manifest.remove.%
image.manifest.remove.%:
	@rm -rf ${HOME}/.docker/manifests/docker.io_$(REGISTRY_PREFIX)_$(IMAGE)-$(VERSION)

.PHONY: image.manifest.push.multiarch
image.manifest.push.multiarch: image.push.multiarch $(addprefix image.manifest.push.multiarch., $(IMAGES))

.PHONY: image.manifest.push.multiarch.%
image.manifest.push.multiarch.%:
	@echo "===========> Pushing manifest $* $(VERSION) to $(REGISTRY_PREFIX) and then remove the local manifest list"
	REGISTRY_PREFIX=$(REGISTRY_PREFIX) PLATFROMS="$(PLATFORMS)" IMAGE=$* VERSION=$(VERSION) DOCKER_CLI_EXPERIMENTAL=enabled \
	  $(ONEX_ROOT)/build/lib/create-manifest.sh
