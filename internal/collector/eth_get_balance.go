package collector

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

type EthGetBalance struct {
	rpc     *rpc.Client
	address string
	desc    *prometheus.Desc
}

func NewEthGetBalance(rpc *rpc.Client, address string) *EthGetBalance {
	return &EthGetBalance{
		rpc:     rpc,
		address: address,
		desc: prometheus.NewDesc(
			"eth_get_balance",
			"get balance",
			nil,
			nil,
		),
	}
}

func (collector *EthGetBalance) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *EthGetBalance) Collect(ch chan<- prometheus.Metric) {
	var result hexutil.Big
	if err := collector.rpc.Call(&result, "eth_getBalance", common.HexToAddress(collector.address), "latest"); err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	i := (*big.Int)(&result)
	value := toEther(i)
	valueFloat64, _ := value.Float64()
	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, valueFloat64)
}

// CONVERTS WEI TO ETH
func toEther(o *big.Int) *big.Float {
	// hex -> balanceAsInt (in wei) -> eth
	result, balanceAsInt := big.NewFloat(0), big.NewFloat(0)
	balanceAsInt.SetInt(o)
	result.Mul(big.NewFloat(0.000000000000000001), balanceAsInt)
	return result
}
