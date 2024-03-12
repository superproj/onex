# ==============================================================================
# Makefile helper functions for deploy docker image in a test kubernetes
#


KUBECTL := kubectl
NAMESPACE ?= onex
CONTEXT ?= kind-onex

DEPLOYS=onex-usercenter onex-gateway

.PHONY: deploy.kind
deploy.kind: $(addprefix deploy.deploy., $(addprefix $(PLATFORM)., $(DEPLOYS))) ## Deploy all configured services.

.PHONY: deploy.kind.%
deploy.kind.%: image.build.% ## Deploy a specified service. (Note: Use `make deploy.<service>` to deploy a specific service.)
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM)))) 
	$(eval DEPLOY := $(word 2,$(subst ., ,$*)))
	@echo "===========> Deploying $(REGISTRY_PREFIX)/$(DEPLOY)-$(ARCH):$(VERSION)"
	@$(KUBECTL) -n $(NAMESPACE) --context=$(CONTEXT) set image deployment/$(DEPLOY) $(DEPLOY)=$(REGISTRY_PREFIX)/$(DEPLOY)-$(ARCH):$(VERSION)

.PHONY: deploy.docker
deploy.docker:
	$(ONEX_ROOT)/scripts/installation/install.sh onex::install::docker::install

.PHONY: deploy.docker.uninstall
deploy.docker.uninstall:
	$(ONEX_ROOT)/scripts/installation/install.sh onex::install::docker::uninstall

.PHONY: deploy.sbs
deploy.sbs:
	$(ONEX_ROOT)/scripts/installation/install.sh onex::install::sbs::install

.PHONY: deploy.sbs.uninstall
deploy.sbs.uninstall:
	$(ONEX_ROOT)/scripts/installation/install.sh onex::install::sbs::uninstall
