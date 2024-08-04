# ==============================================================================
# Makefile helper functions for copyright
#
#
.PHONY: copyright.verify
copyright.verify: tools.verify.addlicense ## Verify the boilerplate headers for all files.
	@echo "===========> Verifying the boilerplate headers for all files"
	@addlicense --check -f $(SCRIPTS_DIR)/boilerplate.txt $(ONEX_ROOT) --skip-dirs=third_party,vendor,_output

.PHONY: copyright.add
copyright.add: tools.verify.addlicense ## Add boilerplate headers for all missing files.
	@addlicense -v -f $(SCRIPTS_DIR)/boilerplate.txt $(ONEX_ROOT) --skip-dirs=third_party,vendor,_output
