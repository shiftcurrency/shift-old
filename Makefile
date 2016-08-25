# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gshift gshift-cross evm all test clean
.PHONY: gshift-linux gshift-linux-386 gshift-linux-amd64 gshift-linux-mips64 gshift-linux-mips64le
.PHONY: gshift-linux-arm gshift-linux-arm-5 gshift-linux-arm-6 gshift-linux-arm-7 gshift-linux-arm64
.PHONY: gshift-darwin gshift-darwin-386 gshift-darwin-amd64
.PHONY: gshift-windows gshift-windows-386 gshift-windows-amd64
.PHONY: gshift-android gshift-ios

GOBIN = build/bin
GO ?= latest

gshift:
	build/env.sh go run build/ci.go install ./cmd/gshift
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gshift\" to launch gshift."

evm:
	build/env.sh go run build/ci.go install ./cmd/evm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/evm to start the evm."

all:
	build/env.sh go run build/ci.go install

test: all
	build/env.sh go run build/ci.go test

clean:
	rm -fr build/_workspace/pkg/ Godeps/_workspace/pkg $(GOBIN)/*

# Cross Compilation Targets (xgo)

gshift-cross: gshift-linux gshift-darwin gshift-windows gshift-android gshift-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gshift-*

gshift-linux: gshift-linux-386 gshift-linux-amd64 gshift-linux-arm gshift-linux-mips64 gshift-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gshift-linux-*

gshift-linux-386:
	build/env.sh go run build/ci.go xgo --go=$(GO) --dest=$(GOBIN) --targets=linux/386 -v ./cmd/gshift
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gshift-linux-* | grep 386

gshift-linux-amd64:
	build/env.sh go run build/ci.go xgo --go=$(GO) --dest=$(GOBIN) --targets=linux/amd64 -v ./cmd/gshift
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gshift-linux-* | grep amd64

gshift-linux-arm: gshift-linux-arm-5 gshift-linux-arm-6 gshift-linux-arm-7 gshift-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gshift-linux-* | grep arm

gshift-linux-arm-5:
	build/env.sh go run build/ci.go xgo --go=$(GO) --dest=$(GOBIN) --targets=linux/arm-5 -v ./cmd/gshift
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gshift-linux-* | grep arm-5

gshift-linux-arm-6:
	build/env.sh go run build/ci.go xgo --go=$(GO) --dest=$(GOBIN) --targets=linux/arm-6 -v ./cmd/gshift
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gshift-linux-* | grep arm-6

gshift-linux-arm-7:
	build/env.sh go run build/ci.go xgo --go=$(GO) --dest=$(GOBIN) --targets=linux/arm-7 -v ./cmd/gshift
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gshift-linux-* | grep arm-7

gshift-linux-arm64:
	build/env.sh go run build/ci.go xgo --go=$(GO) --dest=$(GOBIN) --targets=linux/arm64 -v ./cmd/gshift
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gshift-linux-* | grep arm64

gshift-linux-mips64:
	build/env.sh go run build/ci.go xgo --go=$(GO) --dest=$(GOBIN) --targets=linux/mips64 -v ./cmd/gshift
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/gshift-linux-* | grep mips64

gshift-linux-mips64le:
	build/env.sh go run build/ci.go xgo --go=$(GO) --dest=$(GOBIN) --targets=linux/mips64le -v ./cmd/gshift
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/gshift-linux-* | grep mips64le

gshift-darwin: gshift-darwin-386 gshift-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gshift-darwin-*

gshift-darwin-386:
	build/env.sh go run build/ci.go xgo --go=$(GO) --dest=$(GOBIN) --targets=darwin/386 -v ./cmd/gshift
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gshift-darwin-* | grep 386

gshift-darwin-amd64:
	build/env.sh go run build/ci.go xgo --go=$(GO) --dest=$(GOBIN) --targets=darwin/amd64 -v ./cmd/gshift
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gshift-darwin-* | grep amd64

gshift-windows: gshift-windows-386 gshift-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gshift-windows-*

gshift-windows-386:
	build/env.sh go run build/ci.go xgo --go=$(GO) --dest=$(GOBIN) --targets=windows/386 -v ./cmd/gshift
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/gshift-windows-* | grep 386

gshift-windows-amd64:
	build/env.sh go run build/ci.go xgo --go=$(GO) --dest=$(GOBIN) --targets=windows/amd64 -v ./cmd/gshift
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gshift-windows-* | grep amd64

gshift-android:
	build/env.sh go run build/ci.go xgo --go=$(GO) --dest=$(GOBIN) --targets=android-21/aar -v ./cmd/gshift
	@echo "Android cross compilation done:"
	@ls -ld $(GOBIN)/gshift-android-*

gshift-ios:
	build/env.sh go run build/ci.go xgo --go=$(GO) --dest=$(GOBIN) --targets=ios-7.0/framework -v ./cmd/gshift
	@echo "iOS framework cross compilation done:"
	@ls -ld $(GOBIN)/gshift-ios-*
