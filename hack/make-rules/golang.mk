# ==============================================================================
# Makefile helper functions for golang
#

GO := go
# Minimum supported go version.
GO_MINIMUM_VERSION ?= 1.22

ifeq ($(PRJ_SRC_PATH),)
	$(error the variable PRJ_SRC_PATH must be set prior to including golang.mk)
endif
ifeq ($(ONEX_ROOT),)
	$(error the variable ONEX_ROOT must be set prior to including golang.mk)
endif


VERSION_PACKAGE := github.com/superproj/onex/pkg/version
# Check if the tree is dirty.  default to dirty
GIT_TREE_STATE:="dirty"
ifeq (, $(shell git status --porcelain 2>/dev/null))
    GIT_TREE_STATE="clean"
endif
GIT_COMMIT:=$(shell git rev-parse HEAD)

GO_LDFLAGS += \
	-X $(VERSION_PACKAGE).gitVersion=$(VERSION) \
	-X $(VERSION_PACKAGE).gitCommit=$(GIT_COMMIT) \
	-X $(VERSION_PACKAGE).gitTreeState=$(GIT_TREE_STATE) \
	-X $(VERSION_PACKAGE).buildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
	-X main.Version=$(VERSION)
ifneq ($(DLV),)
	GO_BUILD_FLAGS += -gcflags "all=-N -l"
else
	# -s removes symbol information; -w removes DWARF debugging information; The final program cannot be debugged with gdb
	GO_BUILD_FLAGS += -ldflags "-s -w"
endif

# Available cpus for compiling, please refer to https://github.com/caicloud/engineering/issues/8186#issuecomment-518656946 for more info
CPUS := $(shell /bin/bash $(ONEX_ROOT)/hack/read_cpus_available.sh)

# Default golang flags used in build and test
# -p: the number of programs that can be run in parallel
# -ldflags: arguments to pass on each go tool link invocation
GO_BUILD_FLAGS += -p="$(CPUS)" -ldflags "$(GO_LDFLAGS)"

ifeq ($(GOOS),windows)
	GO_OUT_EXT := .exe
endif

GOPATH := $(shell go env GOPATH)
ifeq ($(origin GOBIN), undefined)
	GOBIN := $(GOPATH)/bin
endif

CMD_DIRS := $(wildcard $(ONEX_ROOT)/cmd/*)
# Filter out directories without Go files, as these directories cannot be compiled.
COMMANDS := $(filter-out $(wildcard %.md), $(foreach dir, $(CMD_DIRS), $(if $(wildcard $(dir)/*.go), $(dir),)))
BINS ?= $(foreach cmd,${COMMANDS},$(notdir ${cmd}))

ifeq (${COMMANDS},)
  $(error Could not determine COMMANDS, set ONEX_ROOT or run in source dir)
endif
ifeq (${BINS},)
  $(error Could not determine BINS, set ONEX_ROOT or run in source dir)
endif

EXCLUDE_TESTS=github.com/superproj/onex/pkg/db,manifests

.PHONY: go.build.verify
go.build.verify: ## Verify supported go versions.
ifneq ($(shell $(GO) version|awk -v min=$(GO_MINIMUM_VERSION) '{gsub(/go/,"",$$3);if($$3 >= min){print 0}else{print 1}}'), 0)
	$(error unsupported go version. Please install a go version which is greater than or equal to '$(GO_MINIMUM_VERSION)')
endif

.PHONY: go.build.%
go.build.%: ## Build specified applications with platform, os and arch.
	$(eval COMMAND := $(word 2,$(subst ., ,$*)))
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	#@ONEX_GIT_VERSION=$(VERSION) $(SCRIPTS_DIR)/build.sh $(COMMAND) $(PLATFORM)
	@if grep -q "func main()" $(ONEX_ROOT)/cmd/$(COMMAND)/*.go &>/dev/null; then \
		echo "===========> Building binary $(COMMAND) $(VERSION) for $(OS) $(ARCH)" ; \
		CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(GO_BUILD_FLAGS) \
		-o $(OUTPUT_DIR)/platforms/$(OS)/$(ARCH)/$(COMMAND)$(GO_OUT_EXT) $(PRJ_SRC_PATH)/cmd/$(COMMAND) ; \
	fi

.PHONY: go.build
go.build: $(addprefix go.build., $(addprefix $(PLATFORM)., $(BINS))) ## Build all applications.

.PHONY: go.build.multiarch
go.build.multiarch: $(foreach p,$(PLATFORMS),$(addprefix go.build., $(addprefix $(p)., $(BINS)))) ## Build all applications with all supported arch.

.PHONY: go.test
go.test: tools.verify.go-junit-report ## Run unit test and generate run report.
	@echo "===========> Run unit test"
	@set -o pipefail; $(GO) test -race -cover -coverprofile=$(OUTPUT_DIR)/coverage.out \
		-timeout=10m -shuffle=on -short -v `go list ./...|\
		egrep -v $(subst $(SPACE),'|',$(sort $(EXCLUDE_TESTS)))` 2>&1 | \
		tee >(go-junit-report --set-exit-code >$(OUTPUT_DIR)/report.xml)
	@sed -i '/mock_.*.go/d' $(OUTPUT_DIR)/coverage.out # remove mock_.*.go files from test coverage
	@$(GO) tool cover -html=$(OUTPUT_DIR)/coverage.out -o $(OUTPUT_DIR)/coverage.html

.PHONY: go.test.cover
go.test.cover: go.test ## Calculate test coverage.
	@$(GO) tool cover -func=$(OUTPUT_DIR)/coverage.out | \
		awk -v target=$(COVERAGE) -f $(SCRIPTS_DIR)/coverage.awk

.PHONY: go.updates
go.updates: tools.verify.go-mod-outdated ## Find outdated dependencies.
	@$(GO) list -u -m -json all | go-mod-outdated -update -direct
