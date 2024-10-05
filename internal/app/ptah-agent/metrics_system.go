package ptah_agent

import (
	"log"
	"syscall"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/loadavg"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/mackerelio/go-osstat/network"
	"github.com/mackerelio/go-osstat/uptime"
	"github.com/prometheus/client_golang/prometheus"
)

var metrics struct {
	diskUsageTotal *prometheus.GaugeVec
	diskUsageFree  *prometheus.GaugeVec
	diskUsageUsed  *prometheus.GaugeVec

	diskIO *prometheus.GaugeVec

	cpuUser   prometheus.Gauge
	cpuSystem prometheus.Gauge
	cpuIdle   prometheus.Gauge
	cpuNice   prometheus.Gauge
	cpuTotal  prometheus.Gauge

	memoryUsed  prometheus.Gauge
	memoryFree  prometheus.Gauge
	memoryTotal prometheus.Gauge

	swapUsed  prometheus.Gauge
	swapFree  prometheus.Gauge
	swapTotal prometheus.Gauge

	loadAvg1  prometheus.Gauge
	loadAvg5  prometheus.Gauge
	loadAvg15 prometheus.Gauge

	networkTxBytes *prometheus.GaugeVec
	networkRxBytes *prometheus.GaugeVec

	uptime prometheus.Gauge
}

func init() {
	namespace := "ptah"
	subsystem := "node"

	metrics.diskUsageTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "disk_total_bytes",
		Help:      "Total disk usage in bytes",
	}, []string{"path"})

	metrics.diskUsageFree = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "disk_free_bytes",
		Help:      "Free disk usage in bytes",
	}, []string{"path"})

	metrics.diskUsageUsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "disk_used_bytes",
		Help:      "Used disk usage in bytes",
	}, []string{"path"})

	metrics.diskIO = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "disk_io_ops_count",
		Help:      "Number of disk operations completed",
	}, []string{"device", "operation"})

	metrics.cpuUser = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "cpu_user",
		Help:      "User CPU time in nanoseconds",
	})

	metrics.cpuSystem = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "cpu_system",
		Help:      "System CPU time in nanoseconds",
	})

	metrics.cpuIdle = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "cpu_idle",
		Help:      "Idle CPU time in nanoseconds",
	})

	metrics.cpuNice = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "cpu_nice",
		Help:      "Nice CPU time in nanoseconds",
	})

	metrics.cpuTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "cpu_total",
		Help:      "Total CPU time in nanoseconds",
	})

	metrics.memoryUsed = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "memory_used_bytes",
		Help:      "Used memory in bytes",
	})

	metrics.memoryFree = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "memory_free_bytes",
		Help:      "Free memory in bytes",
	})

	metrics.memoryTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "memory_total_bytes",
		Help:      "Total memory in bytes",
	})

	metrics.swapUsed = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "swap_used_bytes",
		Help:      "Used swap in bytes",
	})

	metrics.swapFree = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "swap_free_bytes",
		Help:      "Free swap in bytes",
	})

	metrics.swapTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "swap_total_bytes",
		Help:      "Total swap in bytes",
	})

	metrics.loadAvg1 = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "load_avg_1m",
		Help:      "Load average over 1 minute",
	})

	metrics.loadAvg5 = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "load_avg_5m",
		Help:      "Load average over 5 minutes",
	})

	metrics.loadAvg15 = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "load_avg_15m",
		Help:      "Load average over 15 minutes",
	})

	metrics.networkTxBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "network_tx_bytes",
		Help:      "Network transmit bytes",
	}, []string{"interface"})

	metrics.networkRxBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "network_rx_bytes",
		Help:      "Network receive bytes",
	}, []string{"interface"})

	metrics.uptime = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "uptime_seconds",
		Help:      "System uptime in seconds",
	})

	prometheus.MustRegister(metrics.diskUsageTotal, metrics.diskUsageFree, metrics.diskUsageUsed,
		metrics.diskIO, metrics.cpuUser, metrics.cpuSystem,
		metrics.cpuIdle, metrics.cpuNice, metrics.cpuTotal, metrics.memoryUsed, metrics.memoryFree,
		metrics.memoryTotal, metrics.swapUsed, metrics.swapFree, metrics.swapTotal, metrics.loadAvg1, metrics.loadAvg5, metrics.loadAvg15,
		metrics.networkTxBytes, metrics.networkRxBytes, metrics.uptime)
}

type DiskUsageRaw struct {
	Path  string
	Total uint64
	Free  uint64
	Used  uint64
}

type DiskIOStatsRaw struct {
	Name            string
	ReadsCompleted  uint64
	WritesCompleted uint64
}

func getDiskUsage(path string) (*DiskUsageRaw, error) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return nil, err
	}
	total := fs.Blocks * uint64(fs.Bsize)
	free := fs.Bfree * uint64(fs.Bsize)
	used := total - free
	return &DiskUsageRaw{
		Path:  path,
		Total: total,
		Free:  free,
		Used:  used,
	}, nil
}

func scrapeSystemMetrics() error {
	diskUsage, err := getDiskUsage("/")
	if err != nil {
		log.Printf("failed to get disk usage: %v", err)
	}

	metrics.diskUsageTotal.WithLabelValues(diskUsage.Path).Set(float64(diskUsage.Total))
	metrics.diskUsageFree.WithLabelValues(diskUsage.Path).Set(float64(diskUsage.Free))
	metrics.diskUsageUsed.WithLabelValues(diskUsage.Path).Set(float64(diskUsage.Used))

	memStats, err := memory.Get()
	if err != nil {
		log.Printf("failed to get memory stats: %v", err)
	}

	metrics.memoryUsed.Set(float64(memStats.Used))
	metrics.memoryFree.Set(float64(memStats.Free))
	metrics.memoryTotal.Set(float64(memStats.Total))
	metrics.swapUsed.Set(float64(memStats.SwapUsed))
	metrics.swapFree.Set(float64(memStats.SwapFree))
	metrics.swapTotal.Set(float64(memStats.SwapTotal))
	cpuStats, err := cpu.Get()
	if err != nil {
		log.Printf("failed to get cpu stats: %v", err)
	}

	metrics.cpuUser.Set(float64(cpuStats.User))
	metrics.cpuSystem.Set(float64(cpuStats.System))
	metrics.cpuIdle.Set(float64(cpuStats.Idle))
	metrics.cpuNice.Set(float64(cpuStats.Nice))
	metrics.cpuTotal.Set(float64(cpuStats.Total))

	diskIOStats, err := getDiskIOStats()
	if err != nil {
		log.Printf("failed to get disk io stats: %v", err)
	}

	for _, stats := range diskIOStats {
		metrics.diskIO.WithLabelValues(stats.Name, "reads").Set(float64(stats.ReadsCompleted))
		metrics.diskIO.WithLabelValues(stats.Name, "writes").Set(float64(stats.WritesCompleted))
	}

	loadAvgStats, err := loadavg.Get()
	if err != nil {
		log.Printf("failed to get load avg stats: %v", err)
	}

	metrics.loadAvg1.Set(loadAvgStats.Loadavg1)
	metrics.loadAvg5.Set(loadAvgStats.Loadavg5)
	metrics.loadAvg15.Set(loadAvgStats.Loadavg15)

	networkStats, err := network.Get()
	if err != nil {
		log.Printf("failed to get network stats: %v", err)
	}

	for _, netStat := range networkStats {
		if netStat.TxBytes > 0 {
			metrics.networkTxBytes.WithLabelValues(netStat.Name).Set(float64(netStat.TxBytes))
		}

		if netStat.RxBytes > 0 {
			metrics.networkRxBytes.WithLabelValues(netStat.Name).Set(float64(netStat.RxBytes))
		}
	}

	uptime, err := uptime.Get()
	if err != nil {
		log.Printf("failed to get uptime: %v", err)
	}

	metrics.uptime.Set(float64(uptime.Seconds()))

	return nil
}
