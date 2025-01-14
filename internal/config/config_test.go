package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseConfigFromFileIsSuccessful(t *testing.T) {
	config, err := ParseConfigFromFile("test_data/test_config.yaml")
	assert.Nil(t, err, "error expected to be nil")

	// General
	assert.Equal(t, "abc", config.General.EthProviderURL)
	assert.Equal(t, "some blockchain name", config.General.EthBlockchainName)
	assert.Equal(t, "qwe", config.General.ServerURL)
	assert.Equal(t, uint64(123), config.General.StartBlockNumber)
	// Targets - ERC-20
	assert.Len(t, config.Target.ERC20, 2)
	assert.Equal(t, "usdt falopa", config.Target.ERC20[0].Name)
	assert.Equal(t, "0x123123", config.Target.ERC20[0].ContractAddr)
	assert.Equal(t, "usdt falopa 2", config.Target.ERC20[1].Name)
	assert.Equal(t, "0x123124", config.Target.ERC20[1].ContractAddr)
	// Targets - Wallets
	assert.Equal(t, "0x123", config.Target.Wallets[0].Addr)
	assert.Equal(t, "wallet 1", config.Target.Wallets[0].Name)
	// Targets - Wallets
	assert.Equal(t, "0x456", config.Target.Wallets[1].Addr)
	assert.Equal(t, "wallet 2", config.Target.Wallets[1].Name)
}

func TestParseConfigFromFileFailsWithNonExistentFile(t *testing.T) {
	config, err := ParseConfigFromFile("abc_test_config.yaml")
	assert.NotNil(t, err)
	assert.Nil(t, config)
}

func TestParseConfigFromFileFailsWithBadFormattedData(t *testing.T) {
	config, err := ParseConfigFromFile("test_data/bad_formatted_config.yaml")
	assert.NotNil(t, err)
	assert.Nil(t, config)
}
