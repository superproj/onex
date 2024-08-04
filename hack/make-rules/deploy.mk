# ==============================================================================
# Makefile helper functions for deploy docker image in a test kubernetes
#


KUBECTL := kubectl
NAMESPACE ?= onex

DEPLOYS ?= onex-usercenter onex-gateway

.PHONY: deploy.run
deploy.run: $(addprefix deploy.run., $(addprefix $(PLATFORM)., $(DEPLOYS))) ## Deploy all configured services.

.PHONY: deploy.run.%
deploy.run.%: image.push.% ## Deploy a specified service.
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM)))) 
	$(eval DEPLOY := $(word 2,$(subst ., ,$*)))
	# In the SemVer versioning specification, the use of the "+" sign is possible, but container image tag names
	# do not support the "+" character. Therefore, it is necessary to replace the "+" with "-" in the version number.
	# For example, the version number "v0.18.0+20240121235656" should be transformed into "v0.18.0-20240121235656" for
	# use as a container tag name.
	$(eval IMAGE_TAG := $(subst +,-,$(VERSION)))
	@echo "===========> Deploying $(REGISTRY_PREFIX)/$(DEPLOY)-$(ARCH):$(IMAGE_TAG)"
	@$(KUBECTL) -n $(NAMESPACE) set image deployment/$(DEPLOY) $(DEPLOY)=$(REGISTRY_PREFIX)/$(DEPLOY)-$(ARCH):$(IMAGE_TAG)

.PHONY: deploy.docker
deploy.docker:
	$(ONEX_ROOT)/hack/installation/install.sh onex::install::docker::install

.PHONY: deploy.docker.uninstall
deploy.docker.uninstall:
	$(ONEX_ROOT)/hack/installation/install.sh onex::install::docker::uninstall

.PHONY: deploy.sbs
deploy.sbs:
	$(ONEX_ROOT)/hack/installation/install.sh onex::install::sbs::install

.PHONY: deploy.sbs.uninstall
deploy.sbs.uninstall:
	$(ONEX_ROOT)/hack/installation/install.sh onex::install::sbs::uninstall
