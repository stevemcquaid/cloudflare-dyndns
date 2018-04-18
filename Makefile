REGISTRY?=stevemcquaid
IMAGE?=cloudflare-dyndns
TEMP_DIR:=$(shell mktemp -d)
ARCH?=amd64
# ALL_ARCH=amd64 arm arm64 ppc64le s390x
ALL_ARCH=amd64
# ML_PLATFORMS=linux/amd64,linux/arm,linux/arm64,linux/ppc64le,linux/s390x
ML_PLATFORMS=linux/amd64
OUT_DIR?=./_output
VENDOR_DOCKERIZED=0
EXEC_NAME=run

VERSION?=latest

ifeq ($(ARCH),amd64)
	BASEIMAGE?=alpine
endif
#ifeq ($(ARCH),arm)
	#BASEIMAGE?=armhf/busybox
#endif
#ifeq ($(ARCH),arm64)
	#BASEIMAGE?=aarch64/busybox
#endif
#ifeq ($(ARCH),ppc64le)
	#BASEIMAGE?=ppc64le/busybox
#endif
#ifeq ($(ARCH),s390x)
	#BASEIMAGE?=s390x/busybox
#endif

.PHONY: all build docker-build push-% push test verify-gofmt gofmt verify

all: build
build: vendor
	CGO_ENABLED=0 GOARCH=$(ARCH) go build -a -tags netgo -o $(OUT_DIR)/$(ARCH)/$(EXEC_NAME) github.com/$(REGISTRY)/$(IMAGE)

docker-build: vendor
	cp deploy/Dockerfile $(TEMP_DIR)
	cd $(TEMP_DIR) && sed -i "s|BASEIMAGE|$(BASEIMAGE)|g" Dockerfile

	docker run -it -v $(TEMP_DIR):/build -v $(shell pwd):/go/src/github.com/$(REGISTRY)/$(IMAGE) -e GOARCH=$(ARCH) golang:1.8 /bin/bash -c "\
		CGO_ENABLED=0 go build -a -tags netgo -o /build/$(EXEC_NAME) github.com/$(REGISTRY)/$(IMAGE)"

	docker build -t $(REGISTRY)/$(IMAGE)-$(ARCH):$(VERSION) $(TEMP_DIR)
	docker tag $(REGISTRY)/$(IMAGE)-$(ARCH):$(VERSION) $(REGISTRY)/$(IMAGE):$(VERSION)  
	rm -rf $(TEMP_DIR)

push-%:
	$(MAKE) ARCH=$* docker-build
	docker push $(REGISTRY)/$(IMAGE)-$*:$(VERSION)

push: ./manifest-tool $(addprefix push-,$(ALL_ARCH))
	./manifest-tool push from-args --platforms $(ML_PLATFORMS) --template $(REGISTRY)/$(IMAGE)-ARCH:$(VERSION) --target $(REGISTRY)/$(IMAGE):$(VERSION)

./manifest-tool:
	curl -sSL https://github.com/estesp/manifest-tool/releases/download/v0.5.0/manifest-tool-linux-amd64 > manifest-tool
	chmod +x manifest-tool

vendor: glide.lock
ifeq ($(VENDOR_DOCKERIZED),1)
	docker run -it -v $(shell pwd):/go/src/github.com/$(REGISTRY)/$(IMAGE) -w /go/src/github.com/$(REGISTRY)/$(IMAGE) golang:1.8 /bin/bash -c "\
		curl https://glide.sh/get | sh \
		&& glide install -v"
else
	glide install -v
endif

test: vendor
	CGO_ENABLED=0 go test ./pkg/...

verify-gofmt:
	./hack/gofmt-all.sh -v

gofmt:
	./hack/gofmt-all.sh

verify: verify-gofmt test

docker-run:
	docker run -it --env-file config.env $(REGISTRY)/$(IMAGE):latest ./$(EXEC_NAME)

run:
	./$(OUT_DIR)/$(ARCH)/$(EXEC_NAME)

deploy-k8s:
	kubectl apply -f deploy/k8s/deployment.yaml
	# kubectl create secret generic $(IMAGE) --from-file=config.env

deploy-k8s-undo:
	kubectl delete -f deploy/k8s/deployment.yaml
	# kubectl delete secret $(IMAGE)