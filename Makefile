.PHONY : default deps build install cert

NAME = registry
HARDWARE = $(shell uname -m)
OS := $(shell uname)
VERSION ?= 0.1.0

default: deps build

deps:
	echo "Configuring Last.Backend"
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

build:
	@echo "== Pre-building configuration"
	mkdir -p build/linux && mkdir -p build/darwin
	@echo "== Building Last.Backend platform"
	@bash ./hack/build-cross.sh

image:
	@echo "== Pre-building configuration"
	@sh ./hack/build-images.sh

install:
	@echo "== Install binaries"
	@bash ./hack/install-cross.sh

cert:
	@echo "== Generate cert"
	mkdir -p cert
	@bash ./hack/ssl/init-ssl-ca cert
	@bash ./hack/ssl/init-ssl cert admin lb-admin IP.1=127.0.0.1
	@bash ./hack/ssl/init-ssl cert builder builder-127.0.0.1 IP.1=127.0.0.1
	@bash ./hack/ssl/init-ssl cert registry registry-127.0.0.1 IP.1=127.0.0.1