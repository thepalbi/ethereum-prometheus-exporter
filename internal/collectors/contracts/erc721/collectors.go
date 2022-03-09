package erc721

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thepalbi/ethereum-prometheus-exporter/clients/erc721"
	"github.com/thepalbi/ethereum-prometheus-exporter/internal/collectors/constants"
	"github.com/thepalbi/ethereum-prometheus-exporter/internal/config"
	"sync"
)

type BlockNumberGetter interface {
	BlockNumber(ctx context.Context) (uint64, error)
}

type ContractClient interface {
	BlockNumberGetter
	bind.ContractFilterer
	bind.ContractCaller
}

type contract struct {
	Address  common.Address
	Symbol   string
	Name     string
	Caller   *erc721.ContractCaller
	Filterer *erc721.ContractFilterer
}

type Event struct {
	contracts        map[string]*contract
	desc             *prometheus.Desc
	collectMutex     sync.Mutex
	lastQueriedBlock uint64
	bnGetter         BlockNumberGetter
}

type TransferEvent struct {
	*Event
}

func NewTransferEvent(client ContractClient, targets []config.ERC721Target, blockNumber uint64, blockchain string) (*TransferEvent, error) {
	contracts := map[string]*contract{}
	for _, target := range targets {
		contract, err := initializeContract(client, target)
		if err != nil {
			return nil, err
		}
		contracts[target.ContractAddr] = contract
	}

	return &TransferEvent{
		&Event{
			desc: prometheus.NewDesc(
				"erc721_transfer_event",
				"ERC721 Transfer events count",
				[]string{"contract", "symbol", constants.NameLabel},
				map[string]string{
					constants.BlockchainNameLabel: blockchain,
				},
			),
			contracts:        contracts,
			bnGetter:         client,
			lastQueriedBlock: blockNumber,
		},
	}, nil
}

func initializeContract(client ContractClient, target config.ERC721Target) (*contract, error) {
	if !common.IsHexAddress(target.ContractAddr) {
		return nil, errors.New("not valid address")
	}
	addr := common.HexToAddress(target.ContractAddr)

	// Initialize actual contract clients
	caller, err := erc721.NewContractCaller(addr, client)
	if err != nil {
		return nil, err
	}
	filterer, err := erc721.NewContractFilterer(addr, client)
	if err != nil {
		return nil, err
	}

	symbol, err := caller.Symbol(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get symbol")
	}

	return &contract{
		Address:  addr,
		Symbol:   symbol,
		Name:     target.Name,
		Caller:   caller,
		Filterer: filterer,
	}, nil
}

func (te *TransferEvent) Describe(ch chan<- *prometheus.Desc) {
	ch <- te.desc
}

func (te *TransferEvent) doCollect(ch chan<- prometheus.Metric, blockNumber uint64, contr *contract) {
	it, err := contr.Filterer.FilterTransfer(&bind.FilterOpts{
		Context: context.Background(),
		Start:   te.lastQueriedBlock,
		End:     &blockNumber,
	}, nil, nil, nil)
	if err != nil {
		wErr := errors.Wrapf(err, "failed to create transfer iterator for contract=[%s]", contr.Address.Hex())
		ch <- prometheus.NewInvalidMetric(te.desc, wErr)
		return
	}

	// histogram summary to collect
	var count uint64 = 0

	for {
		eventsLeft := it.Next()
		if !eventsLeft && it.Error() == nil {
			// Finished reading events, advance lastQueriedBlock and publish histogram data
			ch <- prometheus.MustNewConstHistogram(te.desc, count, 0, nil, contr.Address.Hex(), contr.Symbol, contr.Name)
			return
		} else if !eventsLeft {
			wErr := errors.Wrapf(err, "failed to read transfer event for contract=[%s]", contr.Address.Hex())
			ch <- prometheus.NewInvalidMetric(te.desc, wErr)
			return
		}
		count += 1
	}
}

func (te *TransferEvent) Collect(ch chan<- prometheus.Metric) {
	te.collectMutex.Lock()
	defer te.collectMutex.Unlock()

	blockNumber, err := te.bnGetter.BlockNumber(context.Background())
	if err != nil {
		wErr := errors.Wrap(err, "failed to get block number")
		ch <- prometheus.NewInvalidMetric(te.desc, wErr)
		return
	}
	// INV: currentBlockNum >= lastQueriedblock

	wg := sync.WaitGroup{}
	for _, contr := range te.contracts {
		wg.Add(1)

		go func(contr *contract) {
			defer wg.Done()
			te.doCollect(ch, blockNumber, contr)
		}(contr)
	}

	wg.Wait()

	// Improve error model, this will advance last seen block even if some client fails
	te.lastQueriedBlock = blockNumber
}
