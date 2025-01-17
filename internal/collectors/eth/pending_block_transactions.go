package eth

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thepalbi/ethereum-prometheus-exporter/internal/collectors/constants"
)

type EthPendingBlockTransactions struct {
	rpc  *rpc.Client
	desc *prometheus.Desc
}

func NewEthPendingBlockTransactions(rpc *rpc.Client, blockchain string) *EthPendingBlockTransactions {
	return &EthPendingBlockTransactions{
		rpc: rpc,
		desc: prometheus.NewDesc(
			"eth_pending_block_transactions",
			"the number of transactions in pending block",
			nil,
			map[string]string{constants.BlockchainNameLabel: blockchain},
		),
	}
}

func (collector *EthPendingBlockTransactions) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *EthPendingBlockTransactions) Collect(ch chan<- prometheus.Metric) {
	var result hexutil.Uint64
	if err := collector.rpc.Call(&result, "eth_getBlockTransactionCountByNumber", "pending"); err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	value := float64(result)
	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, value)
}
