package collector

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestEthGetBalance(t *testing.T) {
	rpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"result": "0x17b6d1ef1dff6b88"}"`))
		if err != nil {
			t.Fatalf("could not write a response: %#v", err)
		}
	}))
	defer rpcServer.Close()

	rpc, err := rpc.DialHTTP(rpcServer.URL)
	if err != nil {
		t.Fatalf("rpc connection error: %#v", err)
	}

	collector := NewEthGetBalance(rpc, "0x7A6A59588B8106045303E1923227a2cefbEC2B66")
	ch := make(chan prometheus.Metric, 1)

	collector.Collect(ch)
	close(ch)

	if got := len(ch); got != 1 {
		t.Fatalf("got %v, want 1", got)
	}

	var metric dto.Metric
	for result := range ch {
		if err := result.Write(&metric); err != nil {
			t.Fatalf("expected metric, got %#v", err)
		}
		if got := len(metric.Label); got > 0 {
			t.Fatalf("expected 0 labels, got %d", got)
		}
		if got := *metric.Gauge.Value; got != 1.7087839335643495 {
			t.Fatalf("got %v, want 1.7087839335643495", got)
		}
	}
}
