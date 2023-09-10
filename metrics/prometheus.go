package metrics

import (
	"github.com/prodadidb/gocache/codec"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespaceCache = "cache"
)

var cacheCollector *prometheus.GaugeVec = initCacheCollector(namespaceCache)

// Prometheus represents the prometheus struct for collecting metrics
type Prometheus struct {
	Service      string
	Collector    *prometheus.GaugeVec
	CodecChannel chan codec.CodecInterface
}

func initCacheCollector(namespace string) *prometheus.GaugeVec {
	c := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "collector",
			Namespace: namespace,
			Help:      "This represent the number of items in cache",
		},
		[]string{"service", "store", "metric"},
	)
	return c
}

// NewPrometheus initializes a new prometheus metric instance
func NewPrometheus(service string) *Prometheus {
	prometheus := &Prometheus{
		Service:      service,
		Collector:    cacheCollector,
		CodecChannel: make(chan codec.CodecInterface, 10000),
	}

	go prometheus.recorder()

	return prometheus
}

// Record records a metric in prometheus by specifying the store name, metric name and value
func (m *Prometheus) Record(store, metric string, value float64) {
	m.Collector.WithLabelValues(m.Service, store, metric).Set(value)
}

// Recorder records metrics in prometheus by retrieving values from the codec channel
func (m *Prometheus) recorder() {
	for codec := range m.CodecChannel {
		stats := codec.GetStats()
		storeType := codec.GetStore().GetType()

		m.Record(storeType, "hit_count", float64(stats.Hits))
		m.Record(storeType, "miss_count", float64(stats.Miss))

		m.Record(storeType, "set_success", float64(stats.SetSuccess))
		m.Record(storeType, "set_error", float64(stats.SetError))

		m.Record(storeType, "delete_success", float64(stats.DeleteSuccess))
		m.Record(storeType, "delete_error", float64(stats.DeleteError))

		m.Record(storeType, "invalidate_success", float64(stats.InvalidateSuccess))
		m.Record(storeType, "invalidate_error", float64(stats.InvalidateError))
	}
}

// RecordFromCodec sends the given codec into the codec channel to be read from recorder
func (m *Prometheus) RecordFromCodec(codec codec.CodecInterface) {
	m.CodecChannel <- codec
}
