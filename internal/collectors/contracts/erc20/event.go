package erc20

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thepalbi/ethereum-prometheus-exporter/token"
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
	contractClients  map[*contractInfo]*token.TokenFilterer
	desc             *prometheus.Desc
	collectMutex     sync.Mutex
	lastQueriedBlock uint64
	bnGetter         BlockNumberGetter
}

func getContractInfo(contractAddr common.Address, contractClient bind.ContractCaller, name string) (*contractInfo, error) {
	contractCaller, err := token.NewTokenCaller(contractAddr, contractClient)
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
