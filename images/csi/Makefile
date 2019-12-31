all: build

.PHONY: build

#VERSION ?= $(shell git describe --always --dirty)
VERSION ?= latest
IMAGE_NAME ?= ics-csi-driver
IMAGE_TAG ?= $(IMAGE_NAME):$(VERSION)
IMAGE_TAR ?= /tmp/$(IMAGE_TAG).tar

build:
	docker build -t $(IMAGE_TAG) .

load: build
	docker image save $(IMAGE_TAG) -o $(IMAGE_TAR)
	ctr cri load $(IMAGE_TAR)

