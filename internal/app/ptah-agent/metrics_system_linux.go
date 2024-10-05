package ptah_agent

import (
	"github.com/mackerelio/go-osstat/disk"
)

func getDiskIOStats() ([]*DiskIOStatsRaw, error) {
	diskIOStats, err := disk.Get()
	if err != nil {
		return nil, err
	}

	result := make([]*DiskIOStatsRaw, 0, len(diskIOStats))
	for _, stat := range diskIOStats {
		result = append(result, &DiskIOStatsRaw{
			Name:            stat.Name,
			ReadsCompleted:  stat.ReadsCompleted,
			WritesCompleted: stat.WritesCompleted,
		})
	}
	return result, nil
}
