# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

.PHONY: build clean fmt lint test vet
.DEFAULT: all
all:  fmt vet lint build test 

# Package list
PKGS=$(shell go list  ./...| grep -v /vendor/)

# Resolving binary dependencies for specific targets
GOLINT=$(shell which golint || echo '')

vet:
	@echo "+ $@"
	@go vet  $(PKGS)

fmt:
	@echo "+ $@"
	@test -z "$$(gofmt -s -l cfg 2>&1 | tee /dev/stderr)" || \
	  (echo >&2 "+ please format Go code with 'gofmt -s'" && false) 
	@test -z "$$(gofmt -s -l cmd 2>&1 |  tee /dev/stderr)" || \
	  (echo >&2 "+ please format Go code with 'gofmt -s'" && false) 
	@test -z "$$(gofmt -s -l crypta 2>&1 | tee /dev/stderr)" || \
	  (echo >&2 "+ please format Go code with 'gofmt -s'" && false) 
	@test -z "$$(gofmt -s -l proto 2>&1 |  tee /dev/stderr)" || \
	  (echo >&2 "+ please format Go code with 'gofmt -s'" && false) 


lint:
	@echo "+ $@"
	-$(GOLINT) cfg/... 
	-$(GOLINT) cmd/... 
	-$(GOLINT) crypta/... 
	-$(GOLINT) proto/... 

build: fmt vet lint 
	@echo "+ $@"
	@go build 

test:
	@echo "+ $@"
	go test  $(PKGS) 

clean:
	@echo "+ $@"

