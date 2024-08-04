# Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file.

# ==============================================================================
# Makefile helper functions for release
#

.PHONY: release.run
release.run: tools.verify.uplift release._check ## Generate and commit CHANGELOG, then tag the git repo.
	@uplift release --fetch-all

release._check:
	@if ! git diff --quiet; then \
        echo "Git repository is not clean"; \
        exit 1; \
    fi
