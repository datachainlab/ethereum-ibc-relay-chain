FORGE  ?= forge
ABIGEN ?= docker run -v .:/workspace -w /workspace -it ethereum/client-go:alltools-v1.11.6 abigen

DOCKER := $(shell which docker)

protoVer=0.13.1
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
	$(FORGE) build --config-path ./yui-ibc-solidity/foundry.toml

.PHONY: abigen
abigen: compile
	@mkdir -p ./build/abi ./pkg/contract/ibchandler
	@jq -r '.abi' ./yui-ibc-solidity/out/IBCHandler.sol/IBCHandler.json > ./build/abi/IBCHandler.abi
	@$(ABIGEN) --abi ./build/abi/IBCHandler.abi --pkg ibchandler --out ./pkg/contract/ibchandler/ibchandler.go

.PHONY: proto-gen proto-update-deps
proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh

proto-update-deps:
	@echo "Updating Protobuf dependencies"
	$(DOCKER) run --user 0 --rm -v $(CURDIR)/proto:/workspace --workdir /workspace $(protoImageName) buf mod update
