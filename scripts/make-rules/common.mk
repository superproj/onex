#
# These variables should not need tweaking.
#

# ==============================================================================
# Includes

# include the common make file
ifeq ($(origin ONEX_ROOT),undefined)
ONEX_ROOT :=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
endif

include $(ONEX_ROOT)/scripts/make-rules/common-versions.mk

# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /usr/bin/env bash -o errexit -o pipefail +o nounset
.SHELLFLAGS = -ec

# It's necessary to set the errexit flags for the bash shell.
export SHELLOPTS := errexit

# ==============================================================================
# Build options
#
PRJ_SRC_PATH :=github.com/superproj/onex

COMMA := ,
SPACE :=
SPACE +=

ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR := $(ONEX_ROOT)/_output
$(shell mkdir -p $(OUTPUT_DIR))
endif

ifeq ($(origin LOCALBIN),undefined)
LOCALBIN := $(OUTPUT_DIR)/bin
$(shell mkdir -p $(LOCALBIN))
endif

ifeq ($(origin TOOLS_DIR),undefined)
TOOLS_DIR := $(OUTPUT_DIR)/tools
$(shell mkdir -p $(TOOLS_DIR))
endif

ifeq ($(origin TMP_DIR),undefined)
TMP_DIR := $(OUTPUT_DIR)/tmp
$(shell mkdir -p $(TMP_DIR))
endif

# set the version number. you should not need to do this
# for the majority of scenarios.
ifeq ($(origin VERSION), undefined)
# Current version of the project.
  VERSION := $(shell git describe --tags --always --match='v*')
  ifneq (,$(shell git status --porcelain 2>/dev/null))
    VERSION := $(VERSION)-dirty
  endif
endif

# Minimum test coverage
ifeq ($(origin COVERAGE),undefined)
COVERAGE := 60
endif

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
GOPATH ?= $(shell go env GOPATH)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# The OS must be linux when building docker images
PLATFORMS ?= linux_amd64 linux_arm64
# The OS can be linux/windows/darwin when building binaries
# PLATFORMS ?= darwin_amd64 windows_amd64 linux_amd64 linux_arm64

# Set a specific PLATFORM
ifeq ($(origin PLATFORM), undefined)
	ifeq ($(origin GOOS), undefined)
		GOOS := $(shell go env GOOS)
	endif
	ifeq ($(origin GOARCH), undefined)
		GOARCH := $(shell go env GOARCH)
	endif
	PLATFORM := $(GOOS)_$(GOARCH)
	# Use linux as the default OS when building images
	IMAGE_PLAT := linux_$(GOARCH)
else
	GOOS := $(word 1, $(subst _, ,$(PLATFORM)))
	GOARCH := $(word 2, $(subst _, ,$(PLATFORM)))
	IMAGE_PLAT := $(PLATFORM)
endif

# Makefile settings
#
# We don't need make's built-in rules.
MAKEFLAGS += --no-builtin-rules
ifeq ($(V),1)
  $(warning ***** starting Makefile for goal(s) "$(MAKECMDGOALS)")
  $(warning ***** $(shell date))
else
  # If we're not debugging the Makefile, don't echo recipes.]
  MAKEFLAGS += -s --no-print-directory
endif

# Linux command settings
FIND := find . ! -path './third_party/*' ! -path './vendor/*'
XARGS := xargs --no-run-if-empty

# Helper function to get dependency version from go.mod
get_go_version = $(shell go list -m $1 | awk '{print $$2}')
define go_install
$(info ===========> Installing $(1)@$(2))
$(GO) install $(1)@$(2)
endef

# Copy githook scripts when execute makefile
COPY_GITHOOK:=$(shell cp -f githooks/* .git/hooks/)

# Specify components which need certificate
ifeq ($(origin CERTIFICATES),undefined)
CERTIFICATES=onex-apiserver admin
endif

MANIFESTS_DIR=$(ONEX_ROOT)/manifests
SCRIPTS_DIR=$(ONEX_ROOT)/scripts

# Image build releated variables.
REGISTRY_PREFIX ?= ccr.ccs.tencentyun.com/superproj
GENERATED_DOCKERFILE_DIR=$(ONEX_ROOT)/build/docker

# Kubernetes releated variables.
## Metadata for driving the build lives here.
META_DIR := $(ONEX_ROOT)/.make
GENERATED_FILE_PREFIX := zz_generated.
EXTRA_GENERATE_PKG := k8s.io/api/core/v1
# This controls the verbosity of the build. Higher numbers mean more output.
KUBE_VERBOSE ?= 1

## Misc
CLIENTSET_NAME_VERSIONED := versioned
OUTPUT_PKG := github.com/superproj/onex/pkg/generated
OPENAPI_EXTRA_PACKAGES := k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/version,k8s.io/kubernetes/pkg/apis/core,k8s.io/api/core/v1,k8s.io/api/autoscaling/v1,k8s.io/api/coordination/v1
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"

HADOLINT_FAILURE_THRESHOLD = warning

APIROOT ?= $(ONEX_ROOT)/pkg/api
APISROOT ?= $(ONEX_ROOT)/pkg/apis

