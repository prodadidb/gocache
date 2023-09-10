package metrics_test

import (
	"testing"
	"time"

	"github.com/prodadidb/gocache/codec"
	"github.com/prodadidb/gocache/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mock_codec_interface_test.go -package=metrics_test -source=../codec/interface.go
//go:generate mockgen -destination=mock_store_interface_test.go -package=metrics_test -source=../store/interface.go

func TestNewPrometheus(t *testing.T) {
	// Given
	serviceName := "my-test-service-name"

	// When
	m := metrics.NewPrometheus(serviceName)

	// Then
	assert.IsType(t, new(metrics.Prometheus), m)

	assert.Equal(t, serviceName, m.Service)
	assert.IsType(t, new(prometheus.GaugeVec), m.Collector)
}

func TestRecord(t *testing.T) {
	// Given
	m := metrics.NewPrometheus("my-test-service-name")

	// When
	m.Record("redis", "hit_count", 6)

	// Then
	metric, err := m.Collector.GetMetricWithLabelValues("my-test-service-name", "redis", "hit_count")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v := testutil.ToFloat64(metric)

	assert.Equal(t, float64(6), v)
}

func TestRecordFromCodec(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	redisStore := NewMockStoreInterface(ctrl)
	redisStore.EXPECT().GetType().Return("redis")

	stats := &codec.Stats{
		Hits:              4,
		Miss:              6,
		SetSuccess:        12,
		SetError:          3,
		DeleteSuccess:     8,
		DeleteError:       5,
		InvalidateSuccess: 2,
		InvalidateError:   1,
	}

	testCodec := NewMockCodecInterface(ctrl)
	testCodec.EXPECT().GetStats().Return(stats)
	testCodec.EXPECT().GetStore().Return(redisStore)

	m := metrics.NewPrometheus("my-test-service-name")

	// When
	m.RecordFromCodec(testCodec)

	// Wait for data to be processed
	for len(m.CodecChannel) > 0 {
		time.Sleep(1 * time.Millisecond)
	}

	// Then
	testCases := []struct {
		metricName string
		expected   float64
	}{
		{
			metricName: "hit_count",
			expected:   float64(stats.Hits),
		},
		{
			metricName: "miss_count",
			expected:   float64(stats.Miss),
		},
		{
			metricName: "set_success",
			expected:   float64(stats.SetSuccess),
		},
		{
			metricName: "set_error",
			expected:   float64(stats.SetError),
		},
		{
			metricName: "delete_success",
			expected:   float64(stats.DeleteSuccess),
		},
		{
			metricName: "delete_error",
			expected:   float64(stats.DeleteError),
		},
		{
			metricName: "invalidate_success",
			expected:   float64(stats.InvalidateSuccess),
		},
		{
			metricName: "invalidate_error",
			expected:   float64(stats.InvalidateError),
		},
	}

	for _, tc := range testCases {
		metric, err := m.Collector.GetMetricWithLabelValues("my-test-service-name", "redis", tc.metricName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v := testutil.ToFloat64(metric)

		assert.Equal(t, tc.expected, v)
	}
}
