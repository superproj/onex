# ==============================================================================
# Makefile used to aggregate all makefiles for easy management.
#

include scripts/make-rules/tools.mk # include at second order
include scripts/make-rules/golang.mk
#include scripts/make-rules/generate-k8s.mk
include scripts/make-rules/generate.mk                                     
include scripts/make-rules/image.mk
include scripts/make-rules/chart.mk
include scripts/make-rules/lint.mk
include scripts/make-rules/swagger.mk
include scripts/make-rules/copyright.mk 
include scripts/make-rules/deploy.mk 
include scripts/make-rules/release.mk 
