package erc20

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/thepalbi/ethereum-prometheus-exporter/internal/collectors/constants"
	"github.com/thepalbi/ethereum-prometheus-exporter/internal/config"
	"log"
	"math"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thepalbi/ethereum-prometheus-exporter/token"
)

type ApprovalEvent struct {
	*Event
}

func NewERC20ApprovalEvent(client ContractClient, contractAddresses []config.ERC20Target, nowBlockNumber uint64, blockchain string) (*ApprovalEvent, error) {
	clients := map[*contractInfo]*token.TokenFilterer{}
	for _, contractAddress := range contractAddresses {
		address := common.HexToAddress(contractAddress.ContractAddr)
		filterer, err := token.NewTokenFilterer(address, client)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create ERC20 approval evt collector")
		}
		info, err := getContractInfo(address, client, contractAddress.Name)
		if err != nil {
			return nil, err
		}

		log.Printf("Got info for %s, symbol %s\n", info.Address, info.Symbol)
		clients[info] = filterer
	}

	return &ApprovalEvent{
		&Event{
			contractClients: clients,
			desc: prometheus.NewDesc(
				"erc20_approval_event",
				"ERC20 Approval events count",
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

func (col *ApprovalEvent) Describe(ch chan<- *prometheus.Desc) {
	ch <- col.desc
}

func (col *ApprovalEvent) doCollect(ch chan<- prometheus.Metric, currentBlockNumber uint64, info *contractInfo, client *token.TokenFilterer) {
	it, err := client.FilterApproval(&bind.FilterOpts{
		Context: context.Background(),
		Start:   col.lastQueriedBlock,
		End:     &currentBlockNumber,
	}, nil, nil)
	if err != nil {
		wErr := errors.Wrapf(err, "failed to create approval iterator for contract=[%s]", info.Address)
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
			ch <- prometheus.MustNewConstHistogram(col.desc, count, sum, nil, info.Address, info.Symbol)
			return
		} else if !eventsLeft {
			wErr := errors.Wrapf(err, "failed to read approval event for contract=[%s]", info.Address)
			ch <- prometheus.NewInvalidMetric(col.desc, wErr)
			return
		}
		te := it.Event

		value, _ := new(big.Float).SetInt(te.Tokens).Float64()
		count += 1
		sum += value / math.Pow10(int(info.Decimals))
	}

}

func (col *ApprovalEvent) Collect(ch chan<- prometheus.Metric) {
	col.collectMutex.Lock()
	defer col.collectMutex.Unlock()

	currentBlockNumber, err := col.bnGetter.BlockNumber(context.Background())
	if err != nil {
		wErr := errors.Wrap(err, "failed to get current block number")
		ch <- prometheus.NewInvalidMetric(col.desc, wErr)
		return
	}

	wg := sync.WaitGroup{}
	for contrInfo, client := range col.contractClients {
		wg.Add(1)

		go func(contrInfo *contractInfo, client *token.TokenFilterer) {
			defer wg.Done()
			col.doCollect(ch, currentBlockNumber, contrInfo, client)
		}(contrInfo, client)
	}

	wg.Wait()

	// Improve error model, this will advance last seen block even if some client fails
	col.lastQueriedBlock = currentBlockNumber
}
