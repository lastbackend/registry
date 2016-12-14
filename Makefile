.PHONY : default deps test build install

NAME = registry
HARDWARE = $(shell uname -m)
OS := $(shell uname)
VERSION ?= 0.1.0

default: deps test build

deps:
	echo "Configuring Last.Backend"
	go get -u github.com/tools/godep
	godep restore

test:
	@echo "Testing Last.Backend"
	@sh ./test/test.sh

build:
	echo "Pre-building configuration"
	mkdir -p build/linux && mkdir -p build/darwin
	echo "Building Last.Backend registry"
	GOOS=linux  go build -ldflags "-X main.Version=$(VERSION)" -o build/linux/$(NAME) cmd/registry/registry.go
	GOOS=darwin go build -ldflags "-X main.Version=$(VERSION)" -o build/darwin/$(NAME) cmd/registry/registry.go

install:
	echo "Install Last.Backend, ${OS} version:= ${VERSION}"
ifeq ($(OS),Linux)
	mv build/linux/$(NAME) /usr/local/bin/$(NAME)
endif
ifeq ($(OS) ,Darwin)
	mv build/darwin/$(NAME) /usr/local/bin/$(NAME)
endif
	chmod +x /usr/local/bin/$(NAME)


