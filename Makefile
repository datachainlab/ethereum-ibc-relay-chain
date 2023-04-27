FORGE  ?= forge
ABIGEN ?= docker run -v .:/workspace -w /workspace -it ethereum/client-go:alltools-v1.11.6 abigen

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
