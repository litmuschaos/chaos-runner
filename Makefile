# Makefile for building Litmus and its tools
# Reference Guide - https://www.gnu.org/software/make/manual/make.html

REGISTRY ?= litmuschaos
IMG_NAME ?= chaos-runner
PACKAGE_VERSION ?= release-multiarch-1.9.0
IS_DOCKER_INSTALLED = $(shell which docker >> /dev/null 2>&1; echo $$?)
HOME = $(shell echo $$HOME)
# list only our namespaced directories
PACKAGES = $(shell go list ./... | grep -v '/vendor/')

.PHONY: all
all: godeps format lint build dockerops test

.PHONY: help
help:
	@echo ""
	@echo "Usage:-"
	@echo "\tmake all   -- [default] builds the chaos runner container"
	@echo ""

.PHONY: godeps
godeps:
	@echo ""
	@echo "INFO:\tverifying dependencies for chaos runner build ..."
	@go get -u -v golang.org/x/lint/golint
	@go get -u -v golang.org/x/tools/cmd/goimports
	#@go get -u -v github.com/golang/dep/cmd/dep

_build_check_docker:
	@if [ $(IS_DOCKER_INSTALLED) -eq 1 ]; \
		then echo "" \
		&& echo "ERROR:\tdocker is not installed. Please install it before build." \
		&& echo "" \
		&& exit 1; \
		fi;

.PHONY: deps
deps: _build_check_docker godeps

.PHONY: format
format:
	@echo "------------------"
	@echo "--> Running go fmt"
	@echo "------------------"
	@go fmt $(PACKAGES)

.PHONY: lint
lint:
	@echo "------------------"
	@echo "--> Running golint"
	@echo "------------------"
	@golint $(PACKAGES)
	@echo "------------------"
	@echo "--> Running go vet"
	@echo "------------------"
	@go vet $(PACKAGES)

.PHONY: build  
build:
	@echo "-----------------------------------"
	@echo "--> Building Chaos-runner binary..."
	@echo "-----------------------------------"
	@./build/go-multiarch-build.sh ./bin

.PHONY: gotasks
gotasks: format lint build

.PHONY: test
test:
	@echo "------------------"
	@echo "Aquire YAML for performing tests"
	@echo "------------------"
	@mkdir -p ./build/_output/test;wget -q -N https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/deploy/chaos_crds.yaml -P ./build/_output/test ;wget -q -N https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/deploy/rbac.yaml -P ./build/_output/test;wget -q -N https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/tests/manifest/pod_delete_rbac.yaml -P ./build/_output/test;wget -q -N https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/deploy/operator.yaml -P ./build/_output/test
	@echo "------------------"
	@echo "--> Run Go Test"
	@echo "------------------"
	@go test ./... -v -count=1

.PHONY: dockerops
dockerops: 
	@echo "------------------"
	@echo "--> Build Chaos-runner image..." 
	@echo "------------------"
	@docker buildx build --file build/Dockerfile  --progress plane --platform linux/arm64,linux/amd64 --tag $(REGISTRY)/$(IMG_NAME):$(PACKAGE_VERSION) .    

.PHONY: push
push:
	@docker buildx build --file build/Dockerfile  --progress plane ----progress plane --push --platform linux/arm64,linux/amd64 --tag $(REGISTRY)/$(IMG_NAME):$(PACKAGE_VERSION) .

