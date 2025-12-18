package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	DiskUsagePercent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "disk_usage_percent",
			Help: "Disk usage percentage",
		}, []string{"path", "type"})

	DiskUsageUsedBytes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "disk_usage_used_bytes",
			Help: "Disk used space in bytes",
		}, []string{"path", "type"})

	DiskUsageAvailBytes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "disk_usage_avail_bytes",
			Help: "Disk available space in bytes",
		}, []string{"path", "type"})
)

func RecordDiskUsageDetailed(path, diskType string, percent int, usedBytes, availBytes float64) {
	DiskUsagePercent.WithLabelValues(path, diskType).Set(float64(percent))
	DiskUsageUsedBytes.WithLabelValues(path, diskType).Set(usedBytes)
	DiskUsageAvailBytes.WithLabelValues(path, diskType).Set(availBytes)
}
