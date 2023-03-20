SHELL := /bin/bash
ARTIFACTS_DIR := ./artifacts
OSNAME = $(shell go env GOOS)
ARCH = $(shell go env GOARCH)
ARCHS := arm64 amd64
OSES := linux darwin

version=$(shell cat version | tr -d '\n')

include linting.mk

define build_target
	CGO_ENABLED=0 \
	GOOS=$(1) \
	GOARCH=$(2) \
	go build \
	-mod=vendor \
	-ldflags="-s -w "\
	-o $(ARTIFACTS_DIR)/s3-copy-$(1)-$(2) \
	./cmd ;
endef

ifeq ($(OS),Windows_NT)
    OSNAME = windows
else
    UNAME_S := $(shell uname -s)
endif

ifdef os
  OSNAME=$(os)
endif

ifndef GOOS
  GOOS=$(OSNAME)
endif

ifndef GOARCH
  GOARCH=amd64
endif

.PHONY: all
all: unit_test lint build

.PHONY: build
build:
	$(foreach os,$(OSES),$(foreach arch,$(ARCHS),$(call build_target,$(os),$(arch))))

.PHONY: local_build
local_build:
	$(call build_target,$(OSNAME),$(ARCH))

.PHONY: deps
deps:
	go mod vendor

.PHONY: test
test:
	go test -v -parallel=6 -mod=vendor -cover $$(go list ./...)
