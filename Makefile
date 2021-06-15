# Makefile for building Litmus and its tools
# Reference Guide - https://www.gnu.org/software/make/manual/make.html

REGISTRY ?= litmuschaos
IMG_NAME ?= chaos-runner
PACKAGE_VERSION ?= ci
IS_DOCKER_INSTALLED = $(shell which docker >> /dev/null 2>&1; echo $$?)
HOME = $(shell echo $$HOME)

.PHONY: help
help:
	@echo ""
	@echo "Usage:-"
	@echo "\tmake godeps    -- sets up dependencies for image build"
	@echo "\tmake build     -- builds the chaos runner binary"
	@echo "\tmake dockerops -- builds & pushes the chaos runner image"
	@echo ""

.PHONY: all
all: godeps build dockerops test

.PHONY: godeps
godeps:
	@echo ""
	@echo "INFO:\tverifying dependencies for chaos runner build ..."
	@go get -u -v golang.org/x/lint/golint
	@go get -u -v golang.org/x/tools/cmd/goimports

_build_check_docker:
	@if [ $(IS_DOCKER_INSTALLED) -eq 1 ]; \
		then echo "" \
		&& echo "ERROR:\tdocker is not installed. Please install it before build." \
		&& echo "" \
		&& exit 1; \
		fi;

.PHONY: deps
deps: _build_check_docker godeps

.PHONY: build  
build:
	@echo "-----------------------------------"
	@echo "--> Building Chaos-runner binary..."
	@echo "-----------------------------------"
	@./build/go-multiarch-build.sh ./bin


.PHONY: test
test:
	@echo "------------------"
	@echo "--> Run Go Test"
	@echo "------------------"
	@go test ./... -coverprofile=coverage.txt -covermode=atomic -v -count=1

.PHONY: dockerops
dockerops: 
	@echo "------------------"
	@echo "--> Build Chaos-runner image..." 
	@echo "------------------"
	@docker buildx build --file build/Dockerfile  --progress plane --platform linux/arm64,linux/amd64 --tag $(REGISTRY)/$(IMG_NAME):$(PACKAGE_VERSION) .    

.PHONY: dockerops-amd64
dockerops-amd64:
	@echo "--------------------------------------------"
	@echo "--> Build chaos-runner amd-64 docker image"
	@echo "--------------------------------------------"
	sudo docker build --file build/Dockerfile --tag $(REGISTRY)/$(IMG_NAME):$(PACKAGE_VERSION) . --build-arg TARGETARCH=amd64
	@echo "--------------------------------------------"
	@echo "--> Push chaos-runner amd-64 docker image"
	@echo "--------------------------------------------"	
	sudo docker push $(REGISTRY)/$(IMG_NAME):$(PACKAGE_VERSION)	

.PHONY: push
push:
	@docker buildx build --file build/Dockerfile  --progress plane --push --platform linux/arm64,linux/amd64 --tag $(REGISTRY)/$(IMG_NAME):$(PACKAGE_VERSION) .

