all: build

# Get the absolute path and name of the current directory.
PWD := $(abspath .)
BASE_DIR := $(notdir $(PWD))

# BUILD_OUT is the root directory containing the build output.
export BUILD_OUT ?= build

# BIN_OUT is the directory containing the built binaries.
export BIN_OUT ?= $(BUILD_OUT)/bin

# DIST_OUT is the directory containting the distribution packages
export DIST_OUT ?= $(BUILD_OUT)/dist

################################################################################
##                             VERIFY GO VERSION                              ##
################################################################################
# Go 1.11+ required for Go modules.
GO_VERSION_EXP := "go1.11"
GO_VERSION_ACT := $(shell a="$$(go version | awk '{print $$3}')" && test $$(printf '%s\n%s' "$${a}" "$(GO_VERSION_EXP)" | sort | tail -n 1) = "$${a}" && printf '%s' "$${a}")
ifndef GO_VERSION_ACT
$(error Requires Go $(GO_VERSION_EXP)+ for Go module support)
endif
MOD_NAME := $(shell head -n 1 <go.mod | awk '{print $$2}')

################################################################################
##                             VERIFY BUILD PATH                              ##
################################################################################
ifneq (on,$(GO111MODULE))
export GO111MODULE := on
# should not be cloned inside the GOPATH.
GOPATH := $(shell go env GOPATH)
ifeq (/src/$(MOD_NAME),$(subst $(GOPATH),,$(PWD)))
$(warning This project uses Go modules and should not be cloned into the GOPATH)
endif
endif

################################################################################
##                                DEPENDENCIES                                ##
################################################################################
# Verify the dependencies are in place.
.PHONY: deps
deps:
	go mod download && go mod verify

################################################################################
##                                VERSIONS                                    ##
################################################################################
# Ensure the version is injected into the binaries via a linker flag.
export VERSION ?= $(shell git describe --always --dirty)

.PHONY: version
version:
	@echo $(VERSION)

################################################################################
##                                BUILD DIRS                                  ##
################################################################################
.PHONY: build-dirs
build-dirs:
	@mkdir -p $(BIN_OUT)
	@mkdir -p $(DIST_OUT)

################################################################################
##                              BUILD BINARIES                                ##
################################################################################
# Unless otherwise specified the binaries should be built for linux-amd64.
GOOS ?= linux
GOARCH ?= amd64

LDFLAGS := $(shell cat hack/make/ldflags.txt)
LDFLAGS_CSI := $(LDFLAGS) -X "$(MOD_NAME)/pkg/csi/service.version=$(VERSION)"
LDFLAGS_SYNCER := $(LDFLAGS)

# The CSI binary.
CSI_BIN_NAME := icsphere-csi
CSI_BIN := $(BIN_OUT)/$(CSI_BIN_NAME).$(GOOS)_$(GOARCH)
build-csi: $(CSI_BIN)
ifndef CSI_BIN_SRCS
CSI_BIN_SRCS := cmd/$(CSI_BIN_NAME)/main.go go.mod go.sum
CSI_BIN_SRCS += $(addsuffix /*.go,$(shell go list -f '{{ join .Deps "\n" }}' ./cmd/$(CSI_BIN_NAME) | grep $(MOD_NAME) | sed 's~$(MOD_NAME)~.~'))
export CSI_BIN_SRCS
endif
$(CSI_BIN): $(CSI_BIN_SRCS)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags '$(LDFLAGS_CSI)' -o $(abspath $@) $<
	@cp -f $@ images/csi/$(CSI_BIN_NAME)
	@touch $@

# The Syncer binary.
SYNCER_BIN_NAME := syncer
SYNCER_BIN := $(BIN_OUT)/$(SYNCER_BIN_NAME).$(GOOS)_$(GOARCH)
build-syncer: $(SYNCER_BIN)
ifndef SYNCER_BIN_SRCS
SYNCER_BIN_SRCS := cmd/$(SYNCER_BIN_NAME)/main.go go.mod go.sum
SYNCER_BIN_SRCS += $(addsuffix /*.go,$(shell go list -f '{{ join .Deps "\n" }}' ./cmd/$(SYNCER_BIN_NAME) | grep $(MOD_NAME) | sed 's~$(MOD_NAME)~.~'))
export SYNCER_BIN_SRCS
endif
$(SYNCER_BIN): $(SYNCER_BIN_SRCS)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags '$(LDFLAGS_SYNCER)' -o $(abspath $@) $<
	@cp -f $@ images/syncer/$(SYNCER_BIN_NAME)
	@touch $@

# The default build target.
build build-bins: $(CSI_BIN) $(SYNCER_BIN)

################################################################################
##                                 CLEAN                                      ##
################################################################################
.PHONY: clean
clean:
	rm -f $(CSI_BIN) $(SYNCER_BIN) images/csi/$(CSI_BIN_NAME) images/syncer/$(SYNCER_BIN_NAME)
	@$(MAKE) -C images/csi clean
	@$(MAKE) -C images/syncer clean
.PHONY: clean-d
clean-d:
	@find . -name "*.d" -type f -delete

################################################################################
##                                 BUILD IMAGES                               ##
################################################################################
.PHONY: csi-image	
csi-image: build-csi
	@$(MAKE) -C images/csi

.PHONY:	syncer-image
syncer-image: build-syncer
	@$(MAKE) -C images/syncer

.PHONY: images
images: csi-image syncer-image

################################################################################
##                                 LINTING                                    ##
################################################################################
.PHONY: check fmt lint mdlint shellcheck vet
check: fmt lint mdlint shellcheck staticcheck vet

fmt:
	hack/check-format.sh

lint:
	hack/check-lint.sh

mdlint:
	hack/check-mdlint.sh

shellcheck:
	hack/check-shell.sh

staticcheck:
	hack/check-staticcheck.sh

vet:
	hack/check-vet.sh
