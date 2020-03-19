all: build

.PHONY: build

VERSION ?= $(shell git describe --always)
IMAGE ?= ics-csi-syncer
IMAGE_TAG ?= $(IMAGE):$(VERSION)
IMAGE_TAR ?= /tmp/$(IMAGE_TAG).tar

build:
	docker build -t $(IMAGE_TAG) .
	docker tag $(IMAGE_TAG) $(IMAGE):latest
load: build
	docker image save $(IMAGE):latest -o $(IMAGE_TAR)
	ctr cri load $(IMAGE_TAR)

