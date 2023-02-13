
.PHONY: build build-alpine clean test help default


BIN_NAME=k8s-whacky-benchmarks
PROJECT_NAME := k8s-whacky-benchmarks
VERSION := $(shell grep "const Version " version/version.go | sed -E 's/.*"(.+)"$$/\1/')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
IMAGE_NAME := "ehlers320/k8s-whacky-benchmarks"
KIND_VERSION?=v1.24.0
KIND_CLUSTER_NAME?=${PROJECT_NAME}-${KIND_VERSION}
KIND_CONTEXT?=kind-${KIND_CLUSTER_NAME}

default: test

help:
	@echo 'Management commands for k8s-whacky-benchmarks:'
	@echo
	@echo 'Usage:'
	@echo '    make build           Compile the project.'
	@echo '    make get-deps        runs dep ensure, mostly used for ci.'
	@echo '    make build-alpine    Compile optimized for alpine linux.'
	@echo '    make package         Build final docker image with just the go binary inside'
	@echo '    make tag             Tag image created by package with latest, git commit and version'
	@echo '    make test            Run tests on a compiled project.'
	@echo '    make push            Push tagged images to registry'
	@echo '    make clean           Clean the directory tree.'
	@echo

######## KIND SECTION ########
kind: kind-create
kind-create:
	kind create cluster --name ${KIND_CLUSTER_NAME} --image kindest/node:${KIND_VERSION}
kind-context:
	kubectl config use ${KIND_CONTEXT}
kind-delete: kind-context
	kind delete clusters ${KIND_CLUSTER_NAME}
kind-metal: kind-context
kind-load-image: docker-build
	kind load docker-image $(IMAGE_NAME):$(GIT_COMMIT) --name ${KIND_CLUSTER_NAME}
kind-fortio: kind-context
	-kubectl create ns fortio
	kubectl -n fortio apply -f tests/manifests/fortio.yaml
######## KIND SECTION ########

build:
	@echo "building ${BIN_NAME} ${VERSION}"
	@echo "GOPATH=${GOPATH}"
	go build -ldflags "-X github.com/tehlers320/k8s-whacky-benchmarks/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/tehlers320/k8s-whacky-benchmarks/version.BuildDate=${BUILD_DATE}" -o bin/${BIN_NAME}

get-deps:
	go mod tidy

build-alpine:
	@echo "building ${BIN_NAME} ${VERSION}"
	@echo "GOPATH=${GOPATH}"
	go build -ldflags '-w -linkmode external -extldflags "-static" -X github.com/tehlers320/k8s-whacky-benchmarks/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/tehlers320/k8s-whacky-benchmarks/version.BuildDate=${BUILD_DATE}' -o bin/${BIN_NAME}

package:
	@echo "building image ${BIN_NAME} ${VERSION} $(GIT_COMMIT)"
	docker build --build-arg VERSION=${VERSION} --build-arg GIT_COMMIT=$(GIT_COMMIT) -t $(IMAGE_NAME):local .

tag: 
	@echo "Tagging: latest ${VERSION} $(GIT_COMMIT)"
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):$(GIT_COMMIT)
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):${VERSION}
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):latest

push: tag
	@echo "Pushing docker image to registry: latest ${VERSION} $(GIT_COMMIT)"
	docker push $(IMAGE_NAME):$(GIT_COMMIT)
	docker push $(IMAGE_NAME):${VERSION}
	docker push $(IMAGE_NAME):latest

clean:
	@test ! -e bin/${BIN_NAME} || rm bin/${BIN_NAME}

test:
	go test ./...

