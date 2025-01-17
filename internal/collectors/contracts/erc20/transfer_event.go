package erc20

import (
	"context"
	"github.com/thepalbi/ethereum-prometheus-exporter/clients/erc20"
	"math"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thepalbi/ethereum-prometheus-exporter/internal/collectors/constants"
	"github.com/thepalbi/ethereum-prometheus-exporter/internal/config"
)

type TransferEvent struct {
	*Event
}

func NewERC20TransferEvent(client ContractClient, contractAddresses []config.ERC20Target, nowBlockNumber uint64, blockchain string) (*TransferEvent, error) {
	clients, err := getContractClients(client, contractAddresses)
	if err != nil {
		return nil, err
	}

	return &TransferEvent{
		&Event{
			contractClients: clients,
			desc: prometheus.NewDesc(
				"erc20_transfer_event",
				"ERC20 Transfer events count",
				[]string{"contract", "symbol", constants.NameLabel},
				map[string]string{
					constants.BlockchainNameLabel: blockchain,
				},
			),
			lastQueriedBlock: nowBlockNumber,
			bnGetter:         client,
		},
	}, nil
}

func (col *TransferEvent) Describe(ch chan<- *prometheus.Desc) {
	ch <- col.desc
}

func (col *TransferEvent) doCollect(ch chan<- prometheus.Metric, currentBlockNumber uint64, info *contractInfo, client *erc20.ContractFilterer) {
	it, err := client.FilterTransfer(&bind.FilterOpts{
		Context: context.Background(),
		Start:   col.lastQueriedBlock,
		End:     &currentBlockNumber,
	}, nil, nil)
	if err != nil {
		wErr := errors.Wrapf(err, "failed to create transfer iterator for contract=[%s]", info.Address)
		ch <- prometheus.NewInvalidMetric(col.desc, wErr)
		return
	}

	// histogram summary to collect
	var count uint64 = 0
	var sum float64 = 0

	for {
		eventsLeft := it.Next()
		if !eventsLeft && it.Error() == nil {
			// Finished reading events, advance lastQueriedBlock and publish histogram data
			ch <- prometheus.MustNewConstHistogram(col.desc, count, sum, nil, info.Address, info.Symbol, info.Name)
			return
		} else if !eventsLeft {
			wErr := errors.Wrapf(err, "failed to read transfer event for contract=[%s]", info.Address)
			ch <- prometheus.NewInvalidMetric(col.desc, wErr)
			return
		}
		te := it.Event

		value, _ := new(big.Float).SetInt(te.Tokens).Float64()
		count += 1
		sum += value / math.Pow10(int(info.Decimals))
	}
}

func (col *TransferEvent) Collect(ch chan<- prometheus.Metric) {
	col.collectMutex.Lock()
	defer col.collectMutex.Unlock()

	currentBlockNumber, err := col.bnGetter.BlockNumber(context.Background())
	if err != nil {
		wErr := errors.Wrap(err, "failed to get current block number")
		ch <- prometheus.NewInvalidMetric(col.desc, wErr)
		return
	}
	// INV: currentBlockNum >= lastQueriedblock

	wg := sync.WaitGroup{}
	for contrInfo, client := range col.contractClients {
		wg.Add(1)

		go func(contrInfo *contractInfo, client *erc20.ContractFilterer) {
			defer wg.Done()
			col.doCollect(ch, currentBlockNumber, contrInfo, client)
		}(contrInfo, client)
	}

	wg.Wait()

	// Improve error model, this will advance last seen block even if some client fails
	col.lastQueriedBlock = currentBlockNumber
}
