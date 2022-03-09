package net

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thepalbi/ethereum-prometheus-exporter/internal/collectors/constants"
)

type NetPeerCount struct {
	rpc  *rpc.Client
	desc *prometheus.Desc
}

func NewNetPeerCount(rpc *rpc.Client, blockchain string) *NetPeerCount {
	return &NetPeerCount{
		rpc: rpc,
		desc: prometheus.NewDesc(
			"net_peers",
			"number of peers currently connected to the client",
			nil,
			map[string]string{constants.BlockchainNameLabel: blockchain},
		),
	}
}

func (collector *NetPeerCount) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.desc
}

func (collector *NetPeerCount) Collect(ch chan<- prometheus.Metric) {
	var result hexutil.Uint64
	if err := collector.rpc.Call(&result, "net_peerCount"); err != nil {
		ch <- prometheus.NewInvalidMetric(collector.desc, err)
		return
	}

	value := float64(result)
	ch <- prometheus.MustNewConstMetric(collector.desc, prometheus.GaugeValue, value)
}
