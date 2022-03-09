SHELL = /bin/bash

all: docker-compose

docker-compose:
	docker-compose build

mixin: lint-mixin

lint-mixin:
	mixtool lint mixin/mixin.libsonnet

build:
	docker-compose build

start:
	docker-compose up -d

stop:
	docker-compose stop

clean: clean-contract-clients

# Contract clients generation
CONTRACTS_DIR := contracts
CONTRACT_CLIENTS_DIR := clients
GENERATED_ABI_DIR := build
CONTRACTS := $(wildcard $(CONTRACTS_DIR)/*.sol)
CONTRACT_CLIENTS := $(patsubst $(CONTRACTS_DIR)/%.sol, $(CONTRACT_CLIENTS_DIR)/%/client.go, $(CONTRACTS))

contract-clients: $(CONTRACT_CLIENTS)

$(CONTRACT_CLIENTS_DIR)/%/client.go:
	@echo "Generating ABI for $@"
	solc --abi $(patsubst $(CONTRACT_CLIENTS_DIR)/%/client.go,$(CONTRACTS_DIR)/%.sol,$@) -o build --overwrite
	@echo "Compiling ABI for $@"
	abigen --abi=$(patsubst $(CONTRACT_CLIENTS_DIR)/%/client.go,$(GENERATED_ABI_DIR)/%.abi,$@) --pkg=$(patsubst $(CONTRACT_CLIENTS_DIR)/%/client.go,%,$@) --out=$@

# Debug goal. Commented
# debug: $(info $$CONTRACT_CLIENTS is [${CONTRACT_CLIENTS}])

clean-contract-clients:
	rm -rf build