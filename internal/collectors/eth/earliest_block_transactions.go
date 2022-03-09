package eth

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thepalbi/ethereum-prometheus-exporter/internal/collectors/constants"
)

type EthEarliestBlockTransactions struct {
	rpc  *rpc.Client
	desc *prometheus.Desc
}

func NewEthEarliestBlockTransactions(rpc *rpc.Client, blockchain string) *EthEarliestBlockTransactions {
	return &EthEarliestBlockTransactions{
		rpc: rpc,
		desc: prometheus.NewDesc(
			"eth_earliest_block_transactions",
			"the number of transactions in an earliest block",
			nil,
			map[string]string{constants.BlockchainNameLabel: blockchain},
		),
	}
}

func (collector *EthEarliestBlockTransactions) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *EthEarliestBlockTransactions) Collect(ch chan<- prometheus.Metric) {
	var result hexutil.Uint64
	if err := collector.rpc.Call(&result, "eth_getBlockTransactionCountByNumber", "earliest"); err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	value := float64(result)
	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, value)
}
