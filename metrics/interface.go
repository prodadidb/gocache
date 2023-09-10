package metrics

import "github.com/prodadidb/gocache/codec"

//go:generate mockgen -destination=./mock_metrics_interface_test.go -package=metrics_test -source=interface.go

// MetricsInterface represents the metrics interface for all available providers
type MetricsInterface interface {
	RecordFromCodec(codec codec.CodecInterface)
}
