package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type ERC20Target struct {
	Name         string `yaml:"name"`
	ContractAddr string `yaml:"contract"`
}

type WalletTarget struct {
	Addr string `yaml:"address"`
	Name string `yaml:"name"`
}

type Config struct {
	General struct {
		EthProviderURL    string `yaml:"eth_provider_url"`
		EthBlockchainName string `yaml:"eth_blockchain_name"`
		ServerURL         string `yaml:"server_url"`
		StartBlockNumber  uint64 `yaml:"start_block_number"`
	} `yaml:"general"`
	Target struct {
		ERC20   []ERC20Target  `yaml:"erc20"`
		Wallets []WalletTarget `yaml:"wallets"`
	} `yaml:"targets"`
}

func ParseConfigFromFile(path string) (*Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := new(Config)
	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
