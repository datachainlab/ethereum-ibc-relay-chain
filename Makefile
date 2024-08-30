FORGE  ?= forge
ABIGEN ?= docker run -v .:/workspace -w /workspace -it ethereum/client-go:alltools-v1.14.0 abigen

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

.PHONY: dep
dep:
	$(DOCKER) run --rm -v $$PWD:$$PWD -w $$PWD node:20 npm i
	cd ./yui-ibc-solidity && $(DOCKER) run --rm -v $$PWD:$$PWD -w $$PWD node:20 npm i

.PHONY: compile
compile:
	$(FORGE) build
	$(FORGE) build --config-path ./yui-ibc-solidity/foundry.toml

.PHONY: abigen
abigen: compile
	@mkdir -p ./build/abi
	@for a in IBCHandler IIBCChannelUpgradableModule; do \
	  b=$$(echo $$a | tr '[A-Z]' '[a-z]'); \
	  mkdir -p ./pkg/contract/$$b; \
	  jq -r '.abi' ./yui-ibc-solidity/out/$$a.sol/$$a.json > ./build/abi/$$a.abi; \
	  $(ABIGEN) --abi ./build/abi/$$a.abi --pkg $$b --out ./pkg/contract/$$b/$$b.go; \
	done
	@for a in Multicall3 IIBCContractUpgradableModule; do \
	  b=$$(echo $$a | tr '[A-Z]' '[a-z]'); \
	  mkdir -p ./pkg/contract/$$b; \
	  jq -r '.abi' ./out/$$a.sol/$$a.json > ./build/abi/$$a.abi; \
	  $(ABIGEN) --abi ./build/abi/$$a.abi --pkg $$b --out ./pkg/contract/$$b/$$b.go; \
	done

.PHONY: proto-gen proto-update-deps
proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh

proto-update-deps:
	@echo "Updating Protobuf dependencies"
	$(DOCKER) run --user 0 --rm -v $(CURDIR)/proto:/workspace --workdir /workspace $(protoImageName) buf mod update
