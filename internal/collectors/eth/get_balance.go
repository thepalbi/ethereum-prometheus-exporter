package eth

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thepalbi/ethereum-prometheus-exporter/internal/collectors/constants"
	"github.com/thepalbi/ethereum-prometheus-exporter/internal/config"
)

type WalletAddress struct {
	Name    string
	Address common.Address
}

type EthGetBalance struct {
	rpc       *rpc.Client
	addresses []WalletAddress
	desc      *prometheus.Desc
}

func NewEthGetBalance(rpc *rpc.Client, wallets []config.WalletTarget, blockchain string) *EthGetBalance {
	var walletAddresses []WalletAddress
	for _, w := range wallets {
		walletAddresses = append(walletAddresses, WalletAddress{w.Name, common.HexToAddress(w.Addr)})
	}
	return &EthGetBalance{
		rpc:       rpc,
		addresses: walletAddresses,
		desc: prometheus.NewDesc(
			"eth_get_balance",
			"get balance",
			[]string{constants.NameLabel},
			map[string]string{
				constants.BlockchainNameLabel: blockchain,
			},
		),
	}
}

func (collector *EthGetBalance) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *EthGetBalance) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	for _, add := range collector.addresses {
		wg.Add(1)
		go func(add WalletAddress) {
			defer wg.Done()
			var result hexutil.Big
			if err := collector.rpc.Call(&result, "eth_getBalance", add.Address, "latest"); err != nil {
				wErr := errors.Wrap(err, "failed to get Balance")
				ch <- prometheus.NewInvalidMetric(collector.desc, wErr)
				return
			}
			balance, _ := new(big.Float).SetInt(result.ToInt()).Float64()
			ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, balance, add.Name)
		}(add)
	}
	wg.Wait()
}
