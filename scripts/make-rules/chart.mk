# ==============================================================================
# Makefile helper functions for tools
#

HELM := helm
CHARTS_DIR=${ONEX_ROOT}/manifests/installation/charts
CHARTS ?= $(foreach charts,$(filter-out %.md, $(wildcard $(CHARTS_DIR)/*)),$(notdir ${charts}))
ifeq (${CHARTS},)
	$(error Could not determine CHARTS, set ONEX_ROOT or run in source dir)
endif

HELM_REPO ?= https://market-tke.tencentcloudcr.com/chartrepo/onex
HELM_REPO_ACCESS_TOKEN ?= # access token for push chart to helm repo

.PHONY: chart.lint
chart.lint: $(addprefix chart.lint., $(CHARTS)) ## Lint all helm charts.

.PHONY: chart.lint.%
chart.lint.%: ## Lint specified helm chart.
	@$(HELM) lint --quiet --with-subcharts $(CHARTS_DIR)/$*

.PHONY: chart.package
chart.package: $(addprefix chart.package., $(CHARTS)) ## Build helm chart packages.

.PHONY: chart.package.%
chart.package.%:  ## Build specified helm chart package.
	@$(HELM) package $(CHARTS_DIR)/$*

.PHONY: chart.upload
chart.upload: $(addprefix chart.upload., $(CHARTS)) ## Upload helm chart packages to helm repo (Need access token: export HELM_REPO_ACCESS_TOKEN=xxxx).

.PHONY: chart.upload.%
chart.upload.%: chart.package.%  # Upload specified helm chart package.
	@$(HELM) repo add onex $(HELM_REPO)
	@$(HELM) cm-push --access-token $(ACCESS_TOKEN) $*-$(CHART_VERSION).tgz onex
