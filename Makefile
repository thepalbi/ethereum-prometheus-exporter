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

# Contract clients generation
CONTRACTS_DIR := contracts
CONTRACT_CLIENTS_DIR := token
GENERATED_ABI_DIR := build
CONTRACTS := $(wildcard $(CONTRACTS_DIR)/*.sol)
CONTRACT_CLIENTS := $(patsubst $(CONTRACTS_DIR)/%.sol, $(CONTRACT_CLIENTS_DIR)/%.go, $(CONTRACTS))

contract-clients: $(CONTRACT_CLIENTS)

$(CONTRACT_CLIENTS_DIR)/%.go:
	@echo "Generating ABI for $@"
	solc --abi $(patsubst $(CONTRACT_CLIENTS_DIR)/%.go,$(CONTRACTS_DIR)/%.sol,$@) -o build --overwrite
	@echo "Compiling ABI for $@"
	abigen --abi=$(patsubst $(CONTRACT_CLIENTS_DIR)/%.go,$(GENERATED_ABI_DIR)/%.abi,$@) --pkg=token --out=$@