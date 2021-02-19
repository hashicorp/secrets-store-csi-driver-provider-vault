REGISTRY_NAME?=docker.io/hashicorp
IMAGE_NAME=secrets-store-csi-driver-provider-vault
IMAGE_VERSION?=$(shell git tag | tail -1)
IMAGE_TAG=$(REGISTRY_NAME)/$(IMAGE_NAME):$(IMAGE_VERSION)
IMAGE_TAG_LATEST=$(REGISTRY_NAME)/$(IMAGE_NAME):latest
BUILD_DATE=$$(date +%Y-%m-%d-%H:%M)
LDFLAGS?="-X 'github.com/hashicorp/secrets-store-csi-driver-provider-vault/internal/version.BuildVersion=$(IMAGE_VERSION)' \
	-X 'github.com/hashicorp/secrets-store-csi-driver-provider-vault/internal/version.BuildDate=$(BUILD_DATE)' \
	-X 'github.com/hashicorp/secrets-store-csi-driver-provider-vault/internal/version.GoVersion=$(shell go version)' \
	-extldflags "-static""
GOOS?=linux
GOARCH?=amd64
GOLANG_IMAGE?=docker.mirror.hashicorp.services/golang:1.15.7
CI_TEST_ARGS=
ifdef CI
override CI_TEST_ARGS:=--junitfile=$(TEST_RESULTS_DIR)/go-test/results.xml --jsonfile=$(TEST_RESULTS_DIR)/go-test/results.json
endif

.PHONY: all test lint build build-in-docker image e2e-container docker-push e2e-setup e2e-teardown e2e-test clean setup mod

GO111MODULE?=on
export GO111MODULE

all: build

lint:
	golangci-lint run -v --concurrency 2 \
		--disable-all \
		--timeout 10m \
		--enable gofmt \
		--enable gosimple \
		--enable govet

test:
	gotestsum --format=short-verbose $(CI_TEST_ARGS)

build: clean
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build \
		-a -ldflags $(LDFLAGS) \
		-o _output/secrets-store-csi-driver-provider-vault_$(GOOS)_$(GOARCH)_$(IMAGE_VERSION) \
		.

build-in-docker: clean
	mkdir -m 777 _output
	docker run --rm \
		--volume `pwd`:`pwd` \
		--workdir=`pwd` \
		--env GOOS \
		--env GOARCH \
		--env LDFLAGS \
		--env REGISTRY_NAME \
		--env IMAGE_VERSION \
		$(GOLANG_IMAGE) \
		make build

image: build-in-docker
	docker build --build-arg VERSION=$(IMAGE_VERSION) -t $(IMAGE_TAG) .

e2e-container:
	REGISTRY_NAME="e2e" IMAGE_VERSION="latest" make image
	kind load docker-image e2e/secrets-store-csi-driver-provider-vault:latest

docker-push:
	docker push $(IMAGE_TAG)
	docker tag $(IMAGE_TAG) $(IMAGE_TAG_LATEST)
	docker push $(IMAGE_TAG_LATEST)

e2e-setup: e2e-container
	kubectl create namespace csi
	helm install secrets-store-csi-driver https://github.com/kubernetes-sigs/secrets-store-csi-driver/blob/master/charts/secrets-store-csi-driver-0.0.19.tgz?raw=true \
		--wait --timeout=5m \
		--namespace=csi \
		--set linux.image.pullPolicy="IfNotPresent" \
		--set grpcSupportedProviders="azure;gcp;vault"
	helm install vault https://github.com/hashicorp/vault-helm/archive/v0.9.1.tar.gz \
		--wait --timeout=5m \
		--namespace=csi \
		--set injector.enabled=false \
		--set server.dev.enabled=true \
		--set server.image.repository=docker.mirror.hashicorp.services/vault
	kubectl apply --namespace=csi -f test/bats/configs/secrets-store-csi-driver-provider-vault.yaml
	kubectl wait --namespace=csi --for=condition=Ready --timeout=5m pod -l app.kubernetes.io/name=vault
	kubectl wait --namespace=csi --for=condition=Ready --timeout=5m pod -l app=secrets-store-csi-driver-provider-vault

e2e-teardown:
	kubectl delete --namespace=csi --ignore-not-found -f test/bats/configs/secrets-store-csi-driver-provider-vault.yaml
	helm uninstall --namespace=csi vault
	helm uninstall --namespace=csi secrets-store-csi-driver
	kubectl delete --ignore-not-found namespace csi

e2e-test:
	bats test/bats/provider.bats

clean:
	-rm -rf _output

mod:
	@go mod tidy
