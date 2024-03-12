# ==============================================================================
# Makefile helper functions for swagger
#

.PHONY: swagger.run
swagger.run: tools.verify.swagger
	@echo "===========> Generating swagger API docs"
	#@swagger generate spec --scan-models -w $(ONEX_ROOT)/cmd/gen-swagger-type-docs -o $(ONEX_ROOT)/api/swagger/kubernetes.yaml
	@swagger mixin `find $(ONEX_ROOT)/api/openapi -name "*.swagger.json"` \
		-q                                                    \
		--keep-spec-order                                     \
		--format=yaml                                         \
		--ignore-conflicts                                    \
		-o $(ONEX_ROOT)/api/swagger/swagger.yaml
	@echo "Generated at: $(ONEX_ROOT)/api/swagger/swagger.yaml"

.PHONY: swagger.serve
swagger.serve: tools.verify.swagger
	@swagger serve -F=redoc --no-open --port 65534 $(ONEX_ROOT)/api/swagger/swagger.yaml
