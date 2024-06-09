# Build all by default, even if it's not first
.DEFAULT_GOAL := help

.PHONY: all
all: format tidy gen add-copyright lint cover build

# ==============================================================================
# Includes

include scripts/make-rules/common.mk # make sure include common.mk at the first include line
include scripts/make-rules/all.mk

# ==============================================================================
# Usage

define USAGE_OPTIONS

\033[35mOptions:\033[0m
  DBG              Whether to generate debug symbols. Default is 0.
  BINS             The binaries to build. Default is all of cmd.
                   This option is available when using: make build/build.multiarch
                   Example: make build BINS="onex-apiserver onex-miner-controller"
  IMAGES           Backend images to make. Default is all of cmd starting with onex-.
                   This option is available when using: make image/image.multiarch/push/push.multiarch
                   Example: make image.multiarch IMAGES="onex-apiserver onex-miner-controller"
  DEPLOYS          Deploy all configured services.
  REGISTRY_PREFIX  Docker registry prefix. Default is superproj. 
                   Example: make push REGISTRY_PREFIX=ccr.ccs.tencentyun.com/superproj VERSION=v0.1.0
  PLATFORMS        The multiple platforms to build. Default is linux_amd64 and linux_arm64.
                   This option is available when using: make build.multiarch/image.multiarch/push.multiarch
                   Example: make image.multiarch IMAGES="onex-apiserver onex-miner-controller" PLATFORMS="linux_amd64 linux_arm64"
  PLATFORMS        The multiple platforms to build. Default is linux_amd64 and linux_arm64.
                   This option is available when using: make build.multiarch/image.multiarch/push.multiarch
  MULTISTAGE       Set to 1 to build docker images using multi-stage builds. Default is 0.
  VERSION          The version information compiled into binaries.
                   The default is obtained from gsemver or git.
  ALL              When Set to 1, it signifies performing a thorough operation.
                   Such as clean all generated files, install all supported tools, generate all files, and so on.
  V                Set to 1 enable verbose build. Default is 0.
endef
export USAGE_OPTIONS

## --------------------------------------
## Generate / Manifests
## --------------------------------------

##@ Generate

.PHONY: gen
gen: ## Generate CI-related files. Generate all files by specifying `A=1`.
ifeq ($(ALL),1)
	$(MAKE) gen.all
else
	$(MAKE) gen.run
endif

.PHONY: gen-k8s
gen-k8s: ## Generate all necessary kubernetes related files, such as deepcopy files
	@$(ONEX_ROOT)/scripts/update-codegen.sh
	# The following command is old generate way with makefile script.
	# Comment here as a code history.
	# $(MAKE) -s generated.files

.PHONY: protoc
protoc: ## Generate api proto files.
	$(MAKE) gen.protoc

.PHONY: ca
ca: ## Generate CA files for all onex components.
	$(MAKE) gen.ca

## --------------------------------------
## Binaries
## --------------------------------------

##@ Build

.PHONY: build
build: tidy ## Build source code for host platform.
	$(MAKE) go.build

.PHONY: build.multiarch
build.multiarch: ## Build source code for multiple platforms. See option PLATFORMS.
	$(MAKE) go.build.multiarch

.PHONY: image
image: ## Build docker images for host arch.
	$(MAKE) image.build

.PHONY: image.multiarch
image.multiarch: ## Build docker images for multiple platforms. See option PLATFORMS.
	$(MAKE) image.build.multiarch

.PHONY: push
push: ## Build docker images for host arch and push images to registry.
	$(MAKE) image.push

.PHONY: push.multiarch
push.multiarch: ## Build docker images for multiple platforms and push images to registry.
	$(MAKE) image.push.multiarch

## --------------------------------------
## Deploy
## --------------------------------------

##@ Deploy

.PHONY: deploy
deploy: ## Build and push docker images for host arch, and deploy it in kubernetes cluster.
	$(MAKE) deploy.run

.PHONY: docker-install
docker-install: ## Deploy onex with docker.
	$(MAKE) deploy.docker

.PHONY: docker-uninstall
docker-uninstall: ## Deploy onex with docker.
	$(MAKE) deploy.docker.uninstall

.PHONY:	sbs-install
sbs-install: ## Deploy onex step by step.
	$(MAKE) deploy.sbs

.PHONY:	sbs-uninstall
sbs-uninstall: ## Deploy onex step by step.
	$(MAKE) deploy.sbs.uninstall

## --------------------------------------
## Cleanup
## --------------------------------------

##@ Clean

.PHONY: clean
clean: ## Remove all artifacts that are created by building and generaters.
	@echo "===========> Cleaning all build output and generated files"
	@-rm -vrf $(OUTPUT_DIR)
	@-rm -vrf $(META_DIR)
ifeq ($(ALL),1)
	@find $(APIROOT) -type f -regextype posix-extended -regex ".*.swagger.json|.*.pb.go" -delete
	@$(FIND) -type f -name 'zz_generated.*go' -delete
	@$(FIND) -type f -name '*_generated.go' -delete
	@$(FIND) -type f -name 'types_swagger_doc_generated.go' -delete
	@-rm -vrf $(ONEX_ROOT)/pkg/generated
	@-rm -vrf $(GENERATED_DOCKERFILE_DIR)
endif

## --------------------------------------
## Testing
## --------------------------------------

##@ Test

.PHONY: test
test: ## Run unit test.
	$(MAKE) go.test

.PHONY: cover 
cover: ## Run unit test and get test coverage.
	$(MAKE) go.test.cover

## --------------------------------------
## Lint / Verification
## --------------------------------------

##@ Lint and Verify

.PHONY: lint
lint: ## Run CI-related linters. Run all linters by specifying `A=1`.
ifeq ($(ALL),1)
	$(MAKE) lint.run
else
	$(MAKE) lint.ci
endif

.PHONY: apidiff
apidiff: tools.verify.go-apidiff ## Run the go-apidiff to verify any API differences compared with origin/master.
	@go-apidiff master --compare-imports --print-compatible --repo-path=.

## --------------------------------------
## Hack / Tools
## --------------------------------------

##@ Hack and Tools

.PHONY: format
format: tools.verify.goimports tools.verify.gofumpt ## Run CI-related formaters. Run all formaters by specifying `A=1`.
	@echo "===========> Formating codes"
	@$(FIND) -type f -name '*.go' | $(XARGS) gofmt -w
	@$(FIND) -type f -name '*.go' | $(XARGS) gofumpt -w
	@$(FIND) -type f -name '*.go' | $(XARGS) goimports -w -local $(PRJ_SRC_PATH)
	@$(GO) mod edit -fmt
ifeq ($(ALL),1)
	$(MAKE) format.protobuf
endif

.PHONY: format.protobuf
format.protobuf: tools.verify.buf ## Lint protobuf files.
	@for f in $(shell find $(APIROOT) -name *.proto) ; do                  \
	  buf format -w $$f ;                                                  \
	done

.PHONY: add-copyright
add-copyright: ## Ensures source code files have copyright license headers.
	$(MAKE) copyright.add

.PHONY: swagger
#swagger: gen.protoc
swagger: ## Generate and aggregate swagger document.
	@$(MAKE) swagger.run

.PHONY: swagger.serve
serve-swagger: ## Serve swagger spec and docs at 65534.
	@$(MAKE) swagger.serve

.PHONY: tidy
tidy:
	@$(GO) mod tidy

.PHONY: air.%
air.%: tools.verify.air
	@air -build.cmd='make build BINS=onex-$*' -build.bin='$(OUTPUT_DIR)/platforms/$(shell go env GOOS)/$(shell go env GOARCH)/onex-$*'

.PHONY: install-tools
install-tools: ## Install CI-related tools. Install all tools by specifying `A=1`.
	$(MAKE) install.ci
	if [[ "$(A)" == 1 ]]; then                                             \
		$(MAKE) _install.other ;                                            \
	fi

.PHONY: release
release: ## Publish a release on the release branch.
	$(MAKE) release.run

.PHONY: targets
targets: Makefile ## Show all Sub-makefile targets.
	@for mk in `echo $(MAKEFILE_LIST) | sed 's/Makefile //g'`; do echo -e \\n\\033[35m$$mk\\033[0m; awk -F':.*##' '/^[0-9A-Za-z._-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^\$$\([0-9A-Za-z_-]+\):.*?##/ { gsub("_","-", $$1); printf "  \033[36m%-45s\033[0m %s\n", tolower(substr($$1, 3, length($$1)-7)), $$2 }' $$mk;done;

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php
.PHONY: help
help: Makefile ## Display this help info.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<TARGETS> <OPTIONS>\033[0m\n\n\033[35mTargets:\033[0m\n"} /^[0-9A-Za-z._-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^\$$\([0-9A-Za-z_-]+\):.*?##/ { gsub("_","-", $$1); printf "  \033[36m%-45s\033[0m %s\n", tolower(substr($$1, 3, length($$1)-7)), $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' Makefile #$(MAKEFILE_LIST)
	@echo -e "$$USAGE_OPTIONS"
