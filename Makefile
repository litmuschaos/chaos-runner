# Makefile for building Litmus and its tools
# Reference Guide - https://www.gnu.org/software/make/manual/make.html

DOCKER_REGISTRY ?= docker.io
DOCKER_REPO ?= litmuschaos
DOCKER_IMAGE ?= chaos-runner
DOCKER_TAG ?= ci
IS_DOCKER_INSTALLED = $(shell which docker >> /dev/null 2>&1; echo $$?)
HOME = $(shell echo $$HOME)

.PHONY: help
help:
	@echo ""
	@echo "Usage:-"
	@echo "\tmake deps                   -- sets up dependencies for image build"
	@echo "\tmake build-chaos-runner   -- builds multi-arch image"
	@echo "\tmake push-chaos-runner    -- pushes the multi-arch image"
	@echo "\tmake build-amd64            -- builds the amd64 image"
	@echo ""	

.PHONY: all
all: deps unused-package-check build-chaos-runner test

.PHONY: deps
deps: _build_check_docker godeps

_build_check_docker:
	@if [ $(IS_DOCKER_INSTALLED) -eq 1 ]; \
		then echo "" \
		&& echo "ERROR:\tdocker is not installed. Please install it before build." \
		&& echo "" \
		&& exit 1; \
		fi;

.PHONY: godeps
godeps:
	@echo ""
	@echo "INFO:\tverifying dependencies for chaos runner build ..."
	@go get -u -v golang.org/x/lint/golint
	@go get -u -v golang.org/x/tools/cmd/goimports

.PHONY: test
test:
	@echo "------------------"
	@echo "--> Run Go Test"
	@echo "------------------"
	@go test ./... -coverprofile=coverage.txt -covermode=atomic -v -count=1


.PHONY: build-chaos-runner
build-chaos-runner:
	@echo "-------------------------"
	@echo "--> Build chaos-runner image" 
	@echo "-------------------------"
	@docker buildx build --file build/Dockerfile --progress plane  --no-cache --platform linux/arm64,linux/amd64 --tag $(DOCKER_REGISTRY)/$(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: push-chaos-runner
push-chaos-runner:
	@echo "------------------------------"
	@echo "--> Pushing image" 
	@echo "------------------------------"
	@docker buildx build --file build/Dockerfile --progress plane --no-cache --push --platform linux/arm64,linux/amd64 --tag $(DOCKER_REGISTRY)/$(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: build-amd64
build-amd64:
	@echo "--------------------------------------"
	@echo "--> Build chaos-runner image for amd64" 
	@echo "--------------------------------------"
	@docker build -f build/Dockerfile  --no-cache -t $(DOCKER_REGISTRY)/$(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG) .  --build-arg TARGETPLATFORM="linux/amd64"

.PHONY: push-amd64
push-amd64:
	@echo "--------------------------------------"
	@echo "--> Push chaos-runner image for amd64" 
	@echo "--------------------------------------"
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG)


.PHONY: unused-package-check
unused-package-check:
	@echo "------------------"
	@echo "--> Check unused packages for the chaos-runner"
	@echo "------------------"
	@tidy=$$(go mod tidy); \
	if [ -n "$${tidy}" ]; then \
		echo "go mod tidy checking failed!"; echo "$${tidy}"; echo; \
	fi

gofmt-check:
	@echo "------------------"
	@echo "--> Check unused packages for the chaos-runner"
	@echo "------------------"
	@gfmt=$$(gofmt -s -l . | wc -l); \
	if [ "$${gfmt}" -ne 0 ]; then \
		echo "The following files were found to be not go formatted:"; \
   		gofmt -s -l .; \
   		exit 1; \
  	fi
