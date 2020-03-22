all: build

.PHONY: build

CSI_BIN_NAME ?= icsphere-csi
TAG := $(shell git describe --always)
IMAGE := ics-csi-driver
IMAGE_TAG := $(IMAGE):$(TAG)
IMAGE_TAR := $(IMAGE)-$(TAG).tar

build:
	docker build -t $(IMAGE_TAG) .
	docker tag $(IMAGE_TAG) $(IMAGE):latest
	docker image save $(IMAGE):latest -o $(IMAGE_TAR)

.PHONY: clean
clean:
	rm -f $(CSI_BIN_NAME)
	rm -f $(IMAGE_TAR)

.PHONY: rmi
rmi: 
	@docker rmi --no-prune $(IMAGE_TAG) $(IMAGE):latest 2>/dev/null || true

.PHONY: clobber
clobber: rmi
	@docker rmi -f $$(docker images -qf reference=$(IMAGE):*) 2>/dev/null || true

.PHONY: load
load: build
	ctr cri load $(IMAGE_TAR)



