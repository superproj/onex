# ==============================================================================
# Makefile helper functions used to generate kubernetes related files and go functions.
#
# Deprecated: Please use hack/update-codegen.sh. This Makefile script is kept for possible future reference.

# Don't allow an implicit 'all' rule.  This is not a user-facing file.
#ifeq ($(MAKECMDGOALS),)
    #$(error This Makefile requires an explicit rule to be specified)
#endif

# Define variables so `make --warn-undefined-variables` works.
UPDATE_API_KNOWN_VIOLATIONS ?=

# This rule collects all the generated file sets into a single rule.  Other
# rules should depend on this to ensure generated files are rebuilt.
.PHONY: generated.files
generated.files: tools.verify.code-generator gen_prerelease_lifecycle gen_deepcopy gen_defaulter gen_conversion gen_openapi gen_client ## Generate all necessary kubernetes files.

#
# Helper logic to find which directories need codegen as quickly as possible.
#

# This variable holds a list of every directory that contains Go files in this
# project.  Other rules and variables can use this as a starting point to
# reduce filesystem accesses.
ifeq ($(V),1)
    $(warning ***** finding all *.go dirs)
endif
ALL_GO_DIRS := $(shell                                                   \
    $(SCRIPTS_DIR)/cache_go_dirs.sh $(META_DIR)/all_go_dirs.mk  \
)
ifeq ($(V),1)
    $(warning ***** found $(shell echo $(ALL_GO_DIRS) | wc -w) *.go dirs)
endif

# Generate a list of all files that have a `+k8s:` comment-tag.  This will be
# used to derive lists of files/dirs for generation tools.
ifeq ($(V),1)
    $(warning ***** finding all +k8s:|+genclien tags)
endif
ALL_K8S_TAG_FILES := $(shell                             \
    find $(ALL_GO_DIRS) -type f -name \*.go  \
        | xargs grep --color=never -l '^// *+[k8s:|genclient]'  \
)
ifeq ($(V),1)
    $(warning ***** found $(shell echo $(ALL_K8S_TAG_FILES) | wc -w) '+k8s:|+genclient' tagged files)
endif

#
# Code generation logic.
#

# prerelease-lifecycle generation
#
# Any package that wants prerelease-lifecycle functions generated must include a
# comment-tag in column 0 of one file of the form:
#     // +k8s:prerelease-lifecycle-gen=true
#

# The result file, in each pkg, of prerelease-lifecycle generation.
PRERELEASE_LIFECYCLE_BASENAME := $(GENERATED_FILE_PREFIX)prerelease-lifecycle
PRERELEASE_LIFECYCLE_FILENAME := $(PRERELEASE_LIFECYCLE_BASENAME).go

# The tool used to generate prerelease-lifecycle code.
PRERELEASE_LIFECYCLE_GEN := prerelease-lifecycle-gen
PRERELEASE_LIFECYCLE_GEN_TODO := $(PRERELEASE_LIFECYCLE_GEN).todo

# Find all the directories that request prerelease-lifecycle generation.
ifeq ($(V),1)
    $(warning ***** finding all +k8s:prerelease-lifecycle-gen tags)
endif
PRERELEASE_LIFECYCLE_DIRS := $(shell                                                 \
    grep --color=never -l '+k8s:prerelease-lifecycle-gen=true' $(ALL_K8S_TAG_FILES)  \
        | xargs -n1 dirname                                                          \
        | LC_ALL=C sort -u                                                           \
)
ifeq ($(V),1)
    $(warning ***** found $(shell echo $(PRERELEASE_LIFECYCLE_DIRS) | wc -w) +k8s:prerelease-lifecycle-gen tagged dirs)
endif
PRERELEASE_LIFECYCLE_FILES := $(addsuffix /$(PRERELEASE_LIFECYCLE_FILENAME), $(PRERELEASE_LIFECYCLE_DIRS))

# Reset the list of packages that need generation.
$(shell mkdir -p $$(dirname $(META_DIR)/$(PRERELEASE_LIFECYCLE_GEN)))
$(shell rm -f $(META_DIR)/$(PRERELEASE_LIFECYCLE_GEN_TODO))

# This rule aggregates the set of files to generate and then generates them all
# in a single run of the tool.
.PHONY: gen_prerelease_lifecycle
gen_prerelease_lifecycle: $(PRERELEASE_LIFECYCLE_GEN) $(META_DIR)/$(PRERELEASE_LIFECYCLE_GEN_TODO) ## Generate lifecycle generated files.
	if [[ -s $(META_DIR)/$(PRERELEASE_LIFECYCLE_GEN_TODO) ]]; then                  \
	    pkgs=$$(cat $(META_DIR)/$(PRERELEASE_LIFECYCLE_GEN_TODO) | paste -sd, -);   \
	    if [[ "$(V)" == 1 ]]; then                                                  \
	        echo "DBG: running $(PRERELEASE_LIFECYCLE_GEN) for $$pkgs";             \
	    fi;                                                                         \
	    N=$$(cat $(META_DIR)/$(PRERELEASE_LIFECYCLE_GEN_TODO) | wc -l);             \
	    echo "Generating prerelease lifecycle code for $$N targets";                \
	    $(SCRIPTS_DIR)/run-in-gopath.sh $(PRERELEASE_LIFECYCLE_GEN)                 \
	        --v $(KUBE_VERBOSE)                                                     \
	        --logtostderr                                                           \
	        -i "$$pkgs"                                                             \
	        -O $(PRERELEASE_LIFECYCLE_BASENAME)                                     \
			--output-base "${GOPATH}/src"                                           \
			--go-header-file ${SCRIPTS_DIR}/boilerplate.go.txt                      \
	        "$$@";                                                                  \
	fi

# For each dir in PRERELEASE_LIFECYCLE_DIRS, this establishes a dependency between the
# output file and the input files that should trigger a rebuild.
#
# Note that this is a deps-only statement, not a full rule (see below).  This
# has to be done in a distinct step because wildcards don't work in static
# pattern rules.
#
# The 'eval' is needed because this has a different RHS for each LHS, and
# would otherwise produce results that make can't parse.
#$(foreach dir, $(PRERELEASE_LIFECYCLE_DIRS), $(eval                            \
    #$(dir)/$(PRERELEASE_LIFECYCLE_FILENAME): $(GODEPS_$(PRJ_SRC_PATH)/$(dir))  \
#))

# How to regenerate prerelease-lifecycle code.  This is a little slow to run,
# so we batch it up and trigger the batch from the 'generated.files' target.
$(META_DIR)/$(PRERELEASE_LIFECYCLE_GEN_TODO): $(PRERELEASE_LIFECYCLE_FILES)

$(PRERELEASE_LIFECYCLE_FILES): $(PRERELEASE_LIFECYCLE_GEN)
	if [[ "$(V)" == 1 ]]; then                              \
	    echo "DBG: prerelease-lifecycle needed in $(@D):";  \
	    ls -lft --full-time $@ || true;                     \
	    ls -lft --full-time $? || true;                     \
	fi
	echo $(PRJ_SRC_PATH)/$(@D) >> $(META_DIR)/$(PRERELEASE_LIFECYCLE_GEN_TODO)

# How to build the generator tool.  The deps for this are defined in
# the $(GO_PKGDEPS_FILE), above.
#
# A word on the need to touch: This rule might trigger if, for example, a
# non-Go file was added or deleted from a directory on which this depends.
# This target needs to be reconsidered, but Go realizes it doesn't actually
# have to be rebuilt.  In that case, make will forever see the dependency as
# newer than the binary, and try to "rebuild" it over and over.  So we touch
# it, and make is happy.
$(PRERELEASE_LIFECYCLE_GEN): tools.verify.prerelease-lifecycle-gen

# Deep-copy generation
#
# Any package that wants deep-copy functions generated must include a
# comment-tag in column 0 of one file of the form:
#     // +k8s:deepcopy-gen=<VALUE>
#
# The <VALUE> may be one of:
#     generate: generate deep-copy functions into the package
#     register: generate deep-copy functions and register them with a
#               scheme

# The result file, in each pkg, of deep-copy generation.
DEEPCOPY_BASENAME := $(GENERATED_FILE_PREFIX)deepcopy
DEEPCOPY_FILENAME := $(DEEPCOPY_BASENAME).go

# The tool used to generate deep copies.
DEEPCOPY_GEN := deepcopy-gen
DEEPCOPY_GEN_TODO := $(DEEPCOPY_GEN).todo

# Find all the directories that request deep-copy generation.
ifeq ($(V),1)
    $(warning ***** finding all +k8s:deepcopy-gen tags)
endif
DEEPCOPY_DIRS := $(shell                                             \
    grep --color=never -l '+k8s:deepcopy-gen=' $(ALL_K8S_TAG_FILES)  \
        | xargs -n1 dirname                                          \
        | LC_ALL=C sort -u                                           \
)
ifeq ($(V),1)
    $(warning ***** found $(shell echo $(DEEPCOPY_DIRS) | wc -w) +k8s:deepcopy-gen tagged dirs)
endif
DEEPCOPY_FILES := $(addsuffix /$(DEEPCOPY_FILENAME), $(DEEPCOPY_DIRS))

# Reset the list of packages that need generation.
$(shell mkdir -p $$(dirname $(META_DIR)/$(DEEPCOPY_GEN)))
$(shell rm -f $(META_DIR)/$(DEEPCOPY_GEN_TODO))

# This rule aggregates the set of files to generate and then generates them all
# in a single run of the tool.
.PHONY: gen_deepcopy
gen_deepcopy: $(DEEPCOPY_GEN) $(META_DIR)/$(DEEPCOPY_GEN_TODO) ## Generate deepcopy generated files.
	if [[ -s $(META_DIR)/$(DEEPCOPY_GEN_TODO) ]]; then                  \
	    pkgs=$$(cat $(META_DIR)/$(DEEPCOPY_GEN_TODO) | paste -sd, -);   \
	    if [[ "$(V)" == 1 ]]; then                                      \
	        echo "DBG: running $(DEEPCOPY_GEN) for $$pkgs";             \
	    fi;                                                             \
	    N=$$(cat $(META_DIR)/$(DEEPCOPY_GEN_TODO) | wc -l);             \
	    echo "Generating deepcopy code for $$N targets";                \
	    $(SCRIPTS_DIR)/run-in-gopath.sh $(DEEPCOPY_GEN)                 \
	        --v $(KUBE_VERBOSE)                                         \
	        --logtostderr                                               \
	        -i "$$pkgs"                                                 \
	        --bounding-dirs $(PRJ_SRC_PATH),"k8s.io/api"                \
	        -O $(DEEPCOPY_BASENAME)                                     \
			--output-base "${GOPATH}/src"                               \
			--go-header-file ${SCRIPTS_DIR}/boilerplate.go.txt          \
	        "$$@";                                                      \
	fi

# For each dir in DEEPCOPY_DIRS, this establishes a dependency between the
# output file and the input files that should trigger a rebuild.
#
# Note that this is a deps-only statement, not a full rule (see below).  This
# has to be done in a distinct step because wildcards don't work in static
# pattern rules.
#
# The 'eval' is needed because this has a different RHS for each LHS, and
# would otherwise produce results that make can't parse.
#$(foreach dir, $(DEEPCOPY_DIRS), $(eval                            \
    #$(dir)/$(DEEPCOPY_FILENAME): $(GODEPS_$(PRJ_SRC_PATH)/$(dir))  \
#))

# How to regenerate deep-copy code.  This is a little slow to run, so we batch
# it up and trigger the batch from the 'generated.files' target.
$(META_DIR)/$(DEEPCOPY_GEN_TODO): $(DEEPCOPY_FILES)

$(DEEPCOPY_FILES): $(DEEPCOPY_GEN)
	if [[ "$(V)" == 1 ]]; then        \
	    echo "DBG: deepcopy needed in $(@D):";  \
	    ls -lft --full-time $@ || true;         \
	    ls -lft --full-time $? || true;         \
	fi
	echo $(PRJ_SRC_PATH)/$(@D) >> $(META_DIR)/$(DEEPCOPY_GEN_TODO)

# How to build the generator tool.  The deps for this are defined in
# the $(GO_PKGDEPS_FILE), above.
#
# A word on the need to touch: This rule might trigger if, for example, a
# non-Go file was added or deleted from a directory on which this depends.
# This target needs to be reconsidered, but Go realizes it doesn't actually
# have to be rebuilt.  In that case, make will forever see the dependency as
# newer than the binary, and try to "rebuild" it over and over.  So we touch
# it, and make is happy.
$(DEEPCOPY_GEN): tools.verify.deepcopy-gen

# Defaulter generation
#
# Any package that wants defaulter functions generated must include a
# comment-tag in column 0 of one file of the form:
#     // +k8s:defaulter-gen=<VALUE>
#
# The <VALUE> depends on context:
#     on types:
#       true:  always generate a defaulter for this type
#       false: never generate a defaulter for this type
#     on functions:
#       covers: if the function name matches SetDefault_NAME, instructs
#               the generator not to recurse
#     on packages:
#       FIELDNAME: any object with a field of this name is a candidate
#                  for having a defaulter generated

# The result file, in each pkg, of defaulter generation.
DEFAULTER_BASENAME := $(GENERATED_FILE_PREFIX)defaults
DEFAULTER_FILENAME := $(DEFAULTER_BASENAME).go

# The tool used to generate defaulters.
DEFAULTER_GEN := defaulter-gen
DEFAULTER_GEN_TODO := $(DEFAULTER_GEN).todo

# All directories that request any form of defaulter generation.
ifeq ($(V),1)
    $(warning ***** finding all +k8s:defaulter-gen tags)
endif
DEFAULTER_DIRS := $(shell                                            \
    grep --color=never -l '+k8s:defaulter-gen=' $(ALL_K8S_TAG_FILES) \
        | xargs -n1 dirname                                          \
        | LC_ALL=C sort -u                                           \
)
ifeq ($(V),1)
    $(warning ***** found $(shell echo $(DEFAULTER_DIRS) | wc -w) +k8s:defaulter-gen tagged dirs)
endif
DEFAULTER_FILES := $(addsuffix /$(DEFAULTER_FILENAME), $(DEFAULTER_DIRS))
DEFAULTER_EXTRA_PEER_PKGS := \
    $(addprefix $(PRJ_SRC_PATH)/, $(DEFAULTER_DIRS))

# Reset the list of packages that need generation.
$(shell mkdir -p $$(dirname $(META_DIR)/$(DEFAULTER_GEN)))
$(shell rm -f $(META_DIR)/$(DEFAULTER_GEN_TODO))

# This rule aggregates the set of files to generate and then generates them all
# in a single run of the tool.
.PHONY: gen_defaulter
gen_defaulter: $(DEFAULTER_GEN) $(META_DIR)/$(DEFAULTER_GEN_TODO) ## Generate defaulter generated files.
	if [[ -s $(META_DIR)/$(DEFAULTER_GEN_TODO) ]]; then                              \
	    pkgs=$$(cat $(META_DIR)/$(DEFAULTER_GEN_TODO) | paste -sd, -);               \
	    if [[ "$(V)" == 1 ]]; then                                                   \
	        echo "DBG: running $(DEFAULTER_GEN) for $$pkgs";                         \
	    fi;                                                                          \
	    N=$$(cat $(META_DIR)/$(DEFAULTER_GEN_TODO) | wc -l);                         \
	    echo "Generating defaulter code for $$N targets";                            \
	    $(SCRIPTS_DIR)/run-in-gopath.sh $(DEFAULTER_GEN)                             \
	        --v $(KUBE_VERBOSE)                                                      \
	        --logtostderr                                                            \
	        -i "$$pkgs"                                                              \
	        --extra-peer-dirs $$(echo $(DEFAULTER_EXTRA_PEER_PKGS) | sed 's/ /,/g')  \
	        -O $(DEFAULTER_BASENAME)                                                 \
			--output-base "${GOPATH}/src"                                            \
			--go-header-file ${SCRIPTS_DIR}/boilerplate.go.txt                 \
	        "$$@";                                                                   \
	fi

# For each dir in DEFAULTER_DIRS, this establishes a dependency between the
# output file and the input files that should trigger a rebuild.
#
# Note that this is a deps-only statement, not a full rule (see below for that).
#
# The 'eval' is needed because this has a different RHS for each LHS, and
# would otherwise produce results that make can't parse.
#$(foreach dir, $(DEFAULTER_DIRS), $(eval                            \
    #$(dir)/$(DEFAULTER_FILENAME): $(GODEPS_$(PRJ_SRC_PATH)/$(dir))  \
#))

# How to regenerate defaulter code.  This is a little slow to run, so we batch
# it up and trigger the batch from the 'generated.files' target.
$(META_DIR)/$(DEFAULTER_GEN_TODO): $(DEFAULTER_FILES)

$(DEFAULTER_FILES): $(DEFAULTER_GEN)
	if [[ "$(V)" == 1 ]]; then                   \
	    echo "DBG: defaulter needed in $(@D):";  \
	    ls -lft --full-time $@ || true;          \
	    ls -lft --full-time $? || true;          \
	fi
	echo $(PRJ_SRC_PATH)/$(@D) >> $(META_DIR)/$(DEFAULTER_GEN_TODO)

# How to build the generator tool.  The deps for this are defined in
# the $(GO_PKGDEPS_FILE), above.
#
# A word on the need to touch: This rule might trigger if, for example, a
# non-Go file was added or deleted from a directory on which this depends.
# This target needs to be reconsidered, but Go realizes it doesn't actually
# have to be rebuilt.  In that case, make will forever see the dependency as
# newer than the binary, and try to "rebuild" it over and over.  So we touch
# it, and make is happy.
$(DEFAULTER_GEN): tools.verify.defaulter-gen

# Conversion generation

# Any package that wants conversion functions generated into it must
# include one or more comment-tags in its `doc.go` file, of the form:
#     // +k8s:conversion-gen=<INTERNAL_TYPES_DIR>
#
# The INTERNAL_TYPES_DIR is a project-local path to another directory
# which should be considered when evaluating peer types for
# conversions.  An optional additional comment of the form
#     // +k8s:conversion-gen-external-types=<EXTERNAL_TYPES_DIR>
#
# identifies where to find the external types; if there is no such
# comment then the external types are sought in the package where the
# `k8s:conversion` tag is found.
#
# Conversions, in both directions, are generated for every type name
# that is defined in both an internal types package and the external
# types package.
#
# TODO: it might be better in the long term to make peer-types explicit in the
# IDL.

# The result file, in each pkg, of conversion generation.
CONVERSION_BASENAME := $(GENERATED_FILE_PREFIX)conversion
CONVERSION_FILENAME := $(CONVERSION_BASENAME).go

# The tool used to generate conversions.
CONVERSION_GEN := conversion-gen
CONVERSION_GEN_TODO := $(CONVERSION_GEN).todo

# The name of the metadata file listing conversion peers for each pkg.
CONVERSIONS_META := conversions.mk

# All directories that request any form of conversion generation.
ifeq ($(V),1)
    $(warning ***** finding all +k8s:conversion-gen tags)
endif
CONVERSION_DIRS := $(shell                                              \
    grep --color=never '^// *+k8s:conversion-gen=' $(ALL_K8S_TAG_FILES) \
        | cut -f1 -d:                                                   \
        | xargs -n1 dirname                                             \
        | LC_ALL=C sort -u                                              \
)
ifeq ($(V),1)
    $(warning ***** found $(shell echo $(CONVERSION_DIRS) | wc -w) +k8s:conversion-gen tagged dirs)
endif
CONVERSION_FILES := $(addsuffix /$(CONVERSION_FILENAME), $(CONVERSION_DIRS))
CONVERSION_EXTRA_PEER_PKGS := \
    k8s.io/kubernetes/pkg/apis/core \
    k8s.io/kubernetes/pkg/apis/core/v1 \
    k8s.io/api/core/v1
CONVERSION_EXTRA_PKGS := $(addprefix $(PRJ_SRC_PATH)/, $(CONVERSION_DIRS))

# Reset the list of packages that need generation.
$(shell mkdir -p $$(dirname $(META_DIR)/$(CONVERSION_GEN)))
$(shell rm -f $(META_DIR)/$(CONVERSION_GEN_TODO))

# This rule aggregates the set of files to generate and then generates them all
# in a single run of the tool.
.PHONY: gen_conversion
gen_conversion: $(CONVERSION_GEN) $(META_DIR)/$(CONVERSION_GEN_TODO) ## Generate conversion generated files.
	if [[ -s $(META_DIR)/$(CONVERSION_GEN_TODO) ]]; then                                \
	    pkgs=$$(cat $(META_DIR)/$(CONVERSION_GEN_TODO) | paste -sd, -);                 \
	    if [[ "$(V)" == 1 ]]; then                                                      \
	        echo "DBG: running $(CONVERSION_GEN) for $$pkgs";                           \
	    fi;                                                                             \
	    N=$$(cat $(META_DIR)/$(CONVERSION_GEN_TODO) | wc -l);                           \
	    echo "Generating conversion code for $$N targets";                              \
	    $(SCRIPTS_DIR)/run-in-gopath.sh $(CONVERSION_GEN)                               \
	        --extra-peer-dirs $$(echo $(CONVERSION_EXTRA_PEER_PKGS) | sed 's/ /,/g')    \
	        --extra-dirs $$(echo $(CONVERSION_EXTRA_PKGS) | sed 's/ /,/g')              \
	        --v $(KUBE_VERBOSE)                                                         \
	        --logtostderr                                                               \
	        -i "$$pkgs"                                                                 \
	        -O $(CONVERSION_BASENAME)                                                   \
			--output-base "${GOPATH}/src"                                               \
			--go-header-file ${SCRIPTS_DIR}/boilerplate.go.txt                          \
	        "$$@";                                                                      \
	fi

# For each dir in CONVERSION_DIRS, this establishes a dependency between the
# output file and the input files that should trigger a rebuild.
#
# Note that this is a deps-only statement, not a full rule (see below for that).
#
# The 'eval' is needed because this has a different RHS for each LHS, and
# would otherwise produce results that make can't parse.
#$(foreach dir, $(CONVERSION_DIRS), $(eval                            \
    #$(dir)/$(CONVERSION_FILENAME): $(GODEPS_$(PRJ_SRC_PATH)/$(dir))  \
#))

# How to regenerate conversion code.  This is a little slow to run, so we batch
# it up and trigger the batch from the 'generated.files' target.
$(META_DIR)/$(CONVERSION_GEN_TODO): $(CONVERSION_FILES)

$(CONVERSION_FILES): $(CONVERSION_GEN)
	if [[ "$(V)" == 1 ]]; then          \
	    echo "DBG: conversion needed in $(@D):";  \
	    ls -lft --full-time $@ || true;           \
	    ls -lft --full-time $? || true;           \
	fi
	echo $(PRJ_SRC_PATH)/$(@D) >> $(META_DIR)/$(CONVERSION_GEN_TODO)

# How to build the generator tool.  The deps for this are defined in
# the $(GO_PKGDEPS_FILE), above.
#
# A word on the need to touch: This rule might trigger if, for example, a
# non-Go file was added or deleted from a directory on which this depends.
# This target needs to be reconsidered, but Go realizes it doesn't actually
# have to be rebuilt.  In that case, make will forever see the dependency as
# newer than the binary, and try to rebuild it over and over.  So we touch it,
# and make is happy.
$(CONVERSION_GEN): tools.verify.conversion-gen

# Client-go generation
#
# Any package that wants client-go functions generated must include a
# comment-tag in column 0 of one file of the form:
#     // +genclient
#

# The tool used to generate client go.
GEN_CLIENT:= gen-client
GEN_CLIENT_TODO := $(GEN_CLIENT).todo

# Find all the directories that request deep-copy generation.
ifeq ($(V),1)
    $(warning ***** finding all +genclient tags)
endif
CLIENT_DIRS := $(shell                                               \
    grep --color=never -l '+genclient' $(ALL_K8S_TAG_FILES)          \
        | xargs -n1 dirname                                          \
        | LC_ALL=C sort -u                                           \
)
ifeq ($(V),1)
    $(warning ***** found $(shell echo $(CLIENT_DIRS) | wc -w) +genclient tagged dirs)
endif
CLIENT_FILES := $(addsuffix /$(CLIENT_FILENAME), $(CLIENT_DIRS))

# Reset the list of packages that need generation.
$(shell mkdir -p $$(dirname $(META_DIR)/$(GEN_CLIENT_TODO)))
$(shell rm -f $(META_DIR)/$(GEN_CLIENT_TODO))

# This rule aggregates the set of files to generate and then generates them all
# in a single run of the tool.
.PHONY: gen_client
gen_client: $(GEN_CLIENT) $(META_DIR)/$(GEN_CLIENT_TODO) ## Generate client-go generated files.
	if [[ -s $(META_DIR)/$(GEN_CLIENT_TODO) ]]; then                    \
	    pkgs=$$(cat $(META_DIR)/$(GEN_CLIENT_TODO) | paste -sd, -);     \
	    if [[ "$(V)" == 1 ]]; then                                      \
	        echo "DBG: running $(GEN_CLIENT) for $$pkgs";               \
	    fi;                                                             \
	    N=$$(cat $(META_DIR)/$(GEN_CLIENT_TODO) | wc -l);               \
	    echo "Generating client for $$N targets";                       \
		echo "Generating clientset at $(OUTPUT_PKG)/clientset";         \
	    $(SCRIPTS_DIR)/run-in-gopath.sh client-gen                      \
	        --v $(KUBE_VERBOSE)                                         \
	        --logtostderr                                               \
	        --input "$(EXTRA_GENERATE_PKG),$$pkgs"                      \
          --included-types-overrides core/v1/Namespace,core/v1/ConfigMap,core/v1/Event,core/v1/Secret \
	        --clientset-name $(CLIENTSET_NAME_VERSIONED)                \
	        --input-base ""                                             \
	        --output-package $(OUTPUT_PKG)/clientset                    \
          --output-base "${GOPATH}/src"                                 \
          --go-header-file ${SCRIPTS_DIR}/boilerplate.go.txt            \
	        "$$@";                                                      \
			$(SCRIPTS_DIR)/fix-generated.sh;                            \
			echo "Generating lister at $(OUTPUT_PKG)/lister";           \
	    $(SCRIPTS_DIR)/run-in-gopath.sh lister-gen                      \
	        --v $(KUBE_VERBOSE)                                         \
	        --logtostderr                                               \
          -i "$(EXTRA_GENERATE_PKG),$$pkgs"                             \
          --included-types-overrides core/v1/Namespace,core/v1/ConfigMap,core/v1/Event,core/v1/Secret \
	        --output-package $(OUTPUT_PKG)/listers                      \
          --output-base "${GOPATH}/src"                                 \
          --go-header-file ${SCRIPTS_DIR}/boilerplate.go.txt            \
	        "$$@";                                                      \
			echo "Generating informer at $(OUTPUT_PKG)/informer";       \
	    $(SCRIPTS_DIR)/run-in-gopath.sh informer-gen                    \
	        --v $(KUBE_VERBOSE)                                         \
	        --logtostderr                                               \
           -i "$(EXTRA_GENERATE_PKG),$$pkgs"                            \
          --included-types-overrides core/v1/Namespace,core/v1/ConfigMap,core/v1/Event,core/v1/Secret \
          --versioned-clientset-package $(OUTPUT_PKG)/clientset/$(CLIENTSET_NAME_VERSIONED)           \
          --listers-package $(OUTPUT_PKG)/listers                       \
	        --output-package $(OUTPUT_PKG)/informers                    \
          --output-base "${GOPATH}/src"                                 \
          --go-header-file ${SCRIPTS_DIR}/boilerplate.go.txt            \
	        "$$@";                                                      \
	fi

# For each dir in CLIENT_DIRS, this establishes a dependency between the
# output file and the input files that should trigger a rebuild.
#
# Note that this is a deps-only statement, not a full rule (see below).  This
# has to be done in a distinct step because wildcards don't work in static
# pattern rules.
#
# The 'eval' is needed because this has a different RHS for each LHS, and
# would otherwise produce results that make can't parse.
#$(foreach dir, $(CLIENT_DIRS), $(eval                            \
    #$(dir)/$(CLIENT_FILENAME): $(GODEPS_$(PRJ_SRC_PATH)/$(dir))  \
#))

# How to regenerate client-go code. This is a little slow to run, so we batch
# it up and trigger the batch from the 'generated.files' target.
$(META_DIR)/$(GEN_CLIENT_TODO): $(CLIENT_FILES)

.PHONY: $(CLIENT_FILES)
$(CLIENT_FILES): 
	if [[ "$(V)" == 1 ]]; then                   \
	    echo "DBG: client-go needed in $(@D):";  \
	    ls -lft --full-time $@ || true;          \
	    ls -lft --full-time $? || true;          \
	fi
	echo $(PRJ_SRC_PATH)/$(@D) >> $(META_DIR)/$(GEN_CLIENT_TODO)

# A word on the need to touch: This rule might trigger if, for example, a
# non-Go file was added or deleted from a directory on which this depends.
# This target needs to be reconsidered, but Go realizes it doesn't actually
# have to be rebuilt.  In that case, make will forever see the dependency as
# newer than the binary, and try to "rebuild" it over and over.  So we touch
# it, and make is happy.
$(GEN_CLIENT): $(addprefix tools.verify., $(CLIENT_TOOLS))

# OpenAPI generation
#
# Any package that wants open-api functions generated must include a
# comment-tag in column 0 of one file of the form:
#     // +k8s:openapi-gen=true
#

# The result file, in each pkg, of deep-copy generation.
OPENAPI_BASENAME := $(GENERATED_FILE_PREFIX)openapi
OPENAPI_FILENAME := $(OPENAPI_BASENAME).go

# The tool used to generate open apis.
OPENAPI_GEN := openapi-gen
OPENAPI_GEN_TODO := $(OPENAPI_GEN).todo

# Find all the directories that request openapi generation.
ifeq ($(V),1)
    $(warning ***** finding all +k8s:openapi-gen tags)
endif
OPENAPI_DIRS := $(shell                                              \
    grep --color=never -l '+k8s:openapi-gen=' $(ALL_K8S_TAG_FILES)   \
        | xargs -n1 dirname                                          \
        | LC_ALL=C sort -u                                           \
)
ifeq ($(V),1)
    $(warning ***** found $(shell echo $(OPENAPI_DIRS) | wc -w) +k8s:openapi-gen tagged dirs)
endif
OPENAPI_FILES := $(addsuffix /$(OPENAPI_FILENAME), $(OPENAPI_DIRS))

# Reset the list of packages that need generation.
$(shell mkdir -p $$(dirname $(META_DIR)/$(OPENAPI_GEN)))
$(shell rm -f $(META_DIR)/$(OPENAPI_GEN_TODO))

# This rule is the user-friendly entrypoint for openapi generation.
.PHONY: gen_openapi
gen_openapi: $(OPENAPI_GEN) $(META_DIR)/$(OPENAPI_GEN_TODO)
	if [[ -s $(META_DIR)/$(OPENAPI_GEN_TODO) ]]; then                   \
	    pkgs=$$(cat $(META_DIR)/$(OPENAPI_GEN_TODO) | paste -sd, -);    \
	    if [[ "$(V)" == 1 ]]; then                                      \
	        echo "DBG: running $(OPENAPI_GEN) for $$pkgs";              \
	    fi;                                                             \
	    N=$$(cat $(META_DIR)/$(OPENAPI_GEN_TODO) | wc -l);              \
	    echo "Generating openapi code for $$N targets";                 \
	    $(SCRIPTS_DIR)/run-in-gopath.sh $(OPENAPI_GEN)                  \
	        --v $(KUBE_VERBOSE)                                         \
	        --logtostderr                                               \
	        -i "$$pkgs"                                                 \
					-i $(OPENAPI_EXTRA_PACKAGES)                        \
	        --output-package $(OUTPUT_PKG)/openapi                      \
	        -O $(OPENAPI_BASENAME)                                      \
					--output-base "${GOPATH}/src"                       \
					--go-header-file ${SCRIPTS_DIR}/boilerplate.go.txt  \
	        "$$@";                                                      \
	fi

# For each dir in DEEPCOPY_DIRS, this establishes a dependency between the
# output file and the input files that should trigger a rebuild.
#
# Note that this is a deps-only statement, not a full rule (see below).  This
# has to be done in a distinct step because wildcards don't work in static
# pattern rules.
#
# The 'eval' is needed because this has a different RHS for each LHS, and
# would otherwise produce results that make can't parse.
#$(foreach dir, $(OPENAPI_DIRS), $(eval                            \
    #$(dir)/$(OPENAPI_FILENAME): $(GODEPS_$(PRJ_SRC_PATH)/$(dir))  \
#))

# How to regenerate deep-copy code.  This is a little slow to run, so we batch
# it up and trigger the batch from the 'generated.files' target.
$(META_DIR)/$(OPENAPI_GEN_TODO): $(OPENAPI_FILES)

$(OPENAPI_FILES): $(OPENAPI_GEN)
	if [[ "$(V)" == 1 ]]; then                  \
	    echo "DBG: openapi needed in $(@D):";   \
	    ls -lft --full-time $@ || true;         \
	    ls -lft --full-time $? || true;         \
	fi
	echo $(PRJ_SRC_PATH)/$(@D) >> $(META_DIR)/$(OPENAPI_GEN_TODO)

# How to build the generator tool.  The deps for this are defined in
# the $(GO_PKGDEPS_FILE), above.
#
# A word on the need to touch: This rule might trigger if, for example, a
# non-Go file was added or deleted from a directory on which this depends.
# This target needs to be reconsidered, but Go realizes it doesn't actually
# have to be rebuilt.  In that case, make will forever see the dependency as
# newer than the binary, and try to "rebuild" it over and over.  So we touch
# it, and make is happy.
$(OPENAPI_GEN): tools.verify.openapi-gen
