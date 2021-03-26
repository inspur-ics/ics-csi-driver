all: build

.PHONY: build

TAG := $(shell git describe --always)
IMAGE := ics-csi-driver
IMAGE_TAG := $(IMAGE):$(TAG)
IMAGE_TAR := $(IMAGE)-$(TAG).tar
IMAGE_LATEST := docker.io/library/$(IMAGE):latest

build:
	docker build -t $(IMAGE_TAG) .
	docker tag $(IMAGE_TAG) $(IMAGE_LATEST)
	docker image save $(IMAGE_LATEST) -o $(IMAGE_TAR)

.PHONY: clean
clean:
	rm -f $(IMAGE_TAR)

.PHONY: rmi
rmi: 
	@docker rmi --no-prune $(IMAGE_TAG) $(IMAGE_LATEST) 2>/dev/null || true

.PHONY: clobber
clobber: rmi
	@docker rmi -f $$(docker images -qf reference=$(IMAGE):*) 2>/dev/null || true

.PHONY: load
load:
	@[ -f $(IMAGE_TAR) ] || make build
	ctr cri load $(IMAGE_TAR)



