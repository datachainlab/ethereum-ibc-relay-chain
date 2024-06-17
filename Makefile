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
	@for a in IBCHandler Multicall3; do \
	  b=$$(echo $$a | tr '[A-Z]' '[a-z]'); \
	  mkdir -p ./build/abi ./pkg/contract/$$b; \
	  jq -r '.abi' ./yui-ibc-solidity/out/$$a.sol/$$a.json > ./build/abi/$$a.abi; \
	  $(ABIGEN) --abi ./build/abi/$$a.abi --pkg $$b --out ./pkg/contract/$$b/$$b.go; \
	done

.PHONY: proto-gen proto-update-deps
proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh

proto-update-deps:
	@echo "Updating Protobuf dependencies"
	$(DOCKER) run --user 0 --rm -v $(CURDIR)/proto:/workspace --workdir /workspace $(protoImageName) buf mod update
