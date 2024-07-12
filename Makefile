FORGE  ?= forge
ABIGEN ?= docker run -u $$(id -u):$$(id -g) -v .:/workspace -w /workspace -it ethereum/client-go:alltools-v1.14.0 abigen

DOCKER := $(shell which docker)

protoVer=0.14.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --user 0 --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

.PHONY: test
test:
	go test -v ./pkg/...

.PHONY: submodule
submodule:
	git submodule update --init
	cd ./yui-ibc-solidity && npm install

.PHONY: compile
compile:
	cp contract/*.sol ./yui-ibc-solidity/contracts/
	$(FORGE) build --config-path ./yui-ibc-solidity/foundry.toml

.PHONY: abigen
abigen: compile
	@mkdir -p ./build/abi
	@mkdir -p ./pkg/contract/ibchandler
	@mkdir -p ./pkg/contract/multicall3
	@mkdir -p ./pkg/contract/ibcchannelupgradablemodule
	@jq -r '.abi' ./yui-ibc-solidity/out/IBCHandler.sol/IBCHandler.json > ./build/abi/IBCHandler.abi
	@jq -r '.abi' ./yui-ibc-solidity/out/Multicall3.sol/Multicall3.json > ./build/abi/Multicall3.abi
	@jq -r '.abi' ./yui-ibc-solidity/out/IBCChannelUpgradableModule.sol/IIBCChannelUpgradableModule.json > ./build/abi/IIBCChannelUpgradableModule.abi
	@$(ABIGEN) --abi ./build/abi/IBCHandler.abi --pkg ibchandler --out ./pkg/contract/ibchandler/ibchandler.go
	@$(ABIGEN) --abi ./build/abi/Multicall3.abi --pkg multicall3 --out ./pkg/contract/multicall3/multicall3.go
	@$(ABIGEN) --abi ./build/abi/IIBCChannelUpgradableModule.abi --pkg ibcchannelupgradablemodule -out ./pkg/contract/ibcchannelupgradablemodule/ibcchannelupgradablemodule.go

.PHONY: proto-gen proto-update-deps
proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh

proto-update-deps:
	@echo "Updating Protobuf dependencies"
	$(DOCKER) run --user 0 --rm -v $(CURDIR)/proto:/workspace --workdir /workspace $(protoImageName) buf mod update
