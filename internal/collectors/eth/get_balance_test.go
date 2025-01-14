package eth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/thepalbi/ethereum-prometheus-exporter/internal/config"
)

func TestEthGetBalance(t *testing.T) {
	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(fmt.Sprintf(`{"result": "%s"}"`, mockResult)))
		if err != nil {
			t.Fatalf("could not write a response: %#v", err)
		}
	}))
	defer rpcServer.Close()

	rpc, err := rpc.DialHTTP(rpcServer.URL)
	if err != nil {
		t.Fatalf("rpc connection error: %#v", err)
	}

	collector := NewEthGetBalance(rpc, []config.WalletTarget{{Addr: mockWalletAddress, Name: mockWalletName}}, mockBlockchainName)
	ch := make(chan prometheus.Metric, 1)

	collector.Collect(ch)

	// only to wait for goroutine inside Collect method
	time.Sleep(2 * time.Second)

	close(ch)

	if got := len(ch); got != 1 {
		t.Fatalf("got %v, want 1", got)
	}

	var metric dto.Metric
	for result := range ch {
		if err := result.Write(&metric); err != nil {
			t.Fatalf("expected metric, got %#v", err)
		}
		if got := len(metric.Label); got != 2 {
			t.Fatalf("expected 2 label2, got %d", got)
		}
		if got := *metric.Gauge.Value; got != mockExpectedValue {
			t.Fatalf("got %v, want %d", got, mockExpectedValue)
		}
	}
}

func TestEthGetBalanceMultipleWallets(t *testing.T) {
	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(fmt.Sprintf(`{"result": "%s"}"`, mockResult)))
		if err != nil {
			t.Fatalf("could not write a response: %#v", err)
		}
	}))
	defer rpcServer.Close()

	rpc, err := rpc.DialHTTP(rpcServer.URL)
	if err != nil {
		t.Fatalf("rpc connection error: %#v", err)
	}

	collector := NewEthGetBalance(rpc, []config.WalletTarget{
		{Addr: mockWalletAddress, Name: mockWalletName},
		{Addr: mockWallet2Address, Name: mockWallet2Name},
	}, mockBlockchainName)
	ch := make(chan prometheus.Metric, 2)

	collector.Collect(ch)

	// only to wait for goroutine inside Collect method
	time.Sleep(2 * time.Second)

	close(ch)

	if got := len(ch); got != 2 {
		t.Fatalf("got %v, want 2", got)
	}

	var metric dto.Metric
	for result := range ch {
		if err := result.Write(&metric); err != nil {
			t.Fatalf("expected metric, got %#v", err)
		}
		if got := len(metric.Label); got != 2 {
			t.Fatalf("expected 2 label2, got %d", got)
		}
		if got := *metric.Gauge.Value; got != mockExpectedValue {
			t.Fatalf("got %v, want %d", got, mockExpectedValue)
		}
	}
}
