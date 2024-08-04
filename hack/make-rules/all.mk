# ==============================================================================
# Makefile used to aggregate all makefiles for easy management.
#

include hack/make-rules/tools.mk # include at second order
include hack/make-rules/golang.mk
#include hack/make-rules/generate-k8s.mk
include hack/make-rules/generate.mk                                     
include hack/make-rules/image.mk
include hack/make-rules/chart.mk
include hack/make-rules/lint.mk
include hack/make-rules/swagger.mk
include hack/make-rules/copyright.mk 
include hack/make-rules/deploy.mk 
include hack/make-rules/release.mk 
