# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gshift gshift-cross evm all test travis-test-with-coverage xgo clean
.PHONY: gshift-linux gshift-linux-arm gshift-linux-386 gshift-linux-amd64
.PHONY: gshift-darwin gshift-darwin-386 gshift-darwin-amd64
.PHONY: gshift-windows gshift-windows-386 gshift-windows-amd64
.PHONY: gshift-android gshift-android-16 gshift-android-21

GOBIN = build/bin

MODE ?= default
GO ?= latest

gshift:
	build/env.sh go install -v $(shell build/flags.sh) ./cmd/gshift
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gshift\" to launch gshift."

gshift-cross: gshift-linux gshift-darwin gshift-windows gshift-android
	@echo "Full cross compilation done:"
	@ls -l $(GOBIN)/gshift-*

gshift-linux: xgo gshift-linux-arm gshift-linux-386 gshift-linux-amd64
	@echo "Linux cross compilation done:"
	@ls -l $(GOBIN)/gshift-linux-*

gshift-linux-386: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=linux/386 -v $(shell build/flags.sh) ./cmd/gshift
	@echo "Linux 386 cross compilation done:"
	@ls -l $(GOBIN)/gshift-linux-* | grep 386

gshift-linux-amd64: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=linux/amd64 -v $(shell build/flags.sh) ./cmd/gshift
	@echo "Linux amd64 cross compilation done:"
	@ls -l $(GOBIN)/gshift-linux-* | grep amd64

gshift-linux-arm: gshift-linux-arm-5 gshift-linux-arm-6 gshift-linux-arm-7 gshift-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -l $(GOBIN)/gshift-linux-* | grep arm

gshift-linux-arm-5: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=linux/arm-5 -v $(shell build/flags.sh) ./cmd/gshift
	@echo "Linux ARMv5 cross compilation done:"
	@ls -l $(GOBIN)/gshift-linux-* | grep arm-5

gshift-linux-arm-6: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=linux/arm-6 -v $(shell build/flags.sh) ./cmd/gshift
	@echo "Linux ARMv6 cross compilation done:"
	@ls -l $(GOBIN)/gshift-linux-* | grep arm-6

gshift-linux-arm-7: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=linux/arm-7 -v $(shell build/flags.sh) ./cmd/gshift
	@echo "Linux ARMv7 cross compilation done:"
	@ls -l $(GOBIN)/gshift-linux-* | grep arm-7

gshift-linux-arm64: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=linux/arm64 -v $(shell build/flags.sh) ./cmd/gshift
	@echo "Linux ARM64 cross compilation done:"
	@ls -l $(GOBIN)/gshift-linux-* | grep arm64

gshift-darwin: gshift-darwin-386 gshift-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -l $(GOBIN)/gshift-darwin-*

gshift-darwin-386: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=darwin/386 -v $(shell build/flags.sh) ./cmd/gshift
	@echo "Darwin 386 cross compilation done:"
	@ls -l $(GOBIN)/gshift-darwin-* | grep 386

gshift-darwin-amd64: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=darwin/amd64 -v $(shell build/flags.sh) ./cmd/gshift
	@echo "Darwin amd64 cross compilation done:"
	@ls -l $(GOBIN)/gshift-darwin-* | grep amd64

gshift-windows: xgo gshift-windows-386 gshift-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -l $(GOBIN)/gshift-windows-*

gshift-windows-386: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=windows/386 -v $(shell build/flags.sh) ./cmd/gshift
	@echo "Windows 386 cross compilation done:"
	@ls -l $(GOBIN)/gshift-windows-* | grep 386

gshift-windows-amd64: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=windows/amd64 -v $(shell build/flags.sh) ./cmd/gshift
	@echo "Windows amd64 cross compilation done:"
	@ls -l $(GOBIN)/gshift-windows-* | grep amd64

gshift-android: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=android/* -v $(shell build/flags.sh) ./cmd/gshift
	@echo "Android cross compilation done:"
	@ls -l $(GOBIN)/gshift-android-*

gshift-ios: gshift-ios-arm-7 gshift-ios-arm64
	@echo "iOS cross compilation done:"
	@ls -l $(GOBIN)/gshift-ios-*

gshift-ios-arm-7: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=ios/arm-7 -v $(shell build/flags.sh) ./cmd/gshift
	@echo "iOS ARMv7 cross compilation done:"
	@ls -l $(GOBIN)/gshift-ios-* | grep arm-7

gshift-ios-arm64: xgo
	build/env.sh $(GOBIN)/xgo --go=$(GO) --buildmode=$(MODE) --dest=$(GOBIN) --targets=ios-7.0/arm64 -v $(shell build/flags.sh) ./cmd/gshift
	@echo "iOS ARM64 cross compilation done:"
	@ls -l $(GOBIN)/gshift-ios-* | grep arm64

evm:
	build/env.sh $(GOROOT)/bin/go install -v $(shell build/flags.sh) ./cmd/evm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/evm to start the evm."

all:
	build/env.sh go install -v $(shell build/flags.sh) ./...

test: all
	build/env.sh go test ./...

travis-test-with-coverage: all
	build/env.sh build/test-global-coverage.sh

xgo:
	build/env.sh go get github.com/karalabe/xgo

clean:
	rm -fr build/_workspace/pkg/ Godeps/_workspace/pkg $(GOBIN)/*
