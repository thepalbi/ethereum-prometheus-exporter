package erc20

import (
	"context"
	"github.com/thepalbi/ethereum-prometheus-exporter/clients/erc20"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thepalbi/ethereum-prometheus-exporter/internal/config"
)

type BlockNumberGetter interface {
	BlockNumber(ctx context.Context) (uint64, error)
}

type ContractClient interface {
	BlockNumberGetter
	bind.ContractFilterer
	bind.ContractCaller
}

type contractInfo struct {
	Address  string
	Symbol   string
	Decimals uint8
	Name     string
}

type Event struct {
	contractClients  map[*contractInfo]*erc20.TokenFilterer
	desc             *prometheus.Desc
	collectMutex     sync.Mutex
	lastQueriedBlock uint64
	bnGetter         BlockNumberGetter
}

func getContractInfo(contractAddr common.Address, contractClient bind.ContractCaller, name string) (*contractInfo, error) {
	contractCaller, err := erc20.NewTokenCaller(contractAddr, contractClient)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get contract info for %s", contractAddr.Hex())
	}
	symbol, err := contractCaller.Symbol(nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get symbol for %s", contractAddr.Hex())
	}
	decimals, err := contractCaller.Decimals(nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get decimals for %s", contractAddr.Hex())
	}
	return &contractInfo{
		Address:  contractAddr.Hex(),
		Symbol:   symbol,
		Decimals: decimals,
		Name:     name,
	}, nil
}

func getContractClients(client ContractClient, contractAddresses []config.ERC20Target) (map[*contractInfo]*erc20.TokenFilterer, error) {
	clients := map[*contractInfo]*erc20.TokenFilterer{}
	for _, contractAddress := range contractAddresses {
		address := common.HexToAddress(contractAddress.ContractAddr)
		filterer, err := erc20.NewTokenFilterer(address, client)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create ERC20 event collector")
		}
		info, err := getContractInfo(address, client, contractAddress.Name)
		if err != nil {
			return nil, err
		}

		log.Printf("Got info for %s, symbol %s\n", info.Address, info.Symbol)
		clients[info] = filterer
	}

	return clients, nil
}
