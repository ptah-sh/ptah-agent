package ptah_agent

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"

	caddyClient "github.com/ptah-sh/ptah-agent/internal/pkg/caddy-client"
)

type MetricsAgent struct {
	client      *SafeClient
	caddyClient *caddyClient.Client
	interval    time.Duration
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

func NewMetricsAgent(client *SafeClient, caddyClient *caddyClient.Client, interval time.Duration) *MetricsAgent {
	return &MetricsAgent{
		client:      client,
		caddyClient: caddyClient,
		interval:    interval,
		stopChan:    make(chan struct{}),
	}
}

func (m *MetricsAgent) ScrapeMetrics(ctx context.Context) ([]string, error) {
	// TODO: adjust the timestamp value to stick to a resolution of 5 seconds (or whatever else server wants)
	//   something like time.Now().Truncate(5 * time.Second).UnixMilli()
	timestamp := time.Now().UnixMilli()

	caddyMetrics, err := m.caddyClient.GetMetrics(ctx)
	if err != nil {
		log.Printf("failed to get caddy metrics: %v", err)
	}

	err = scrapeSystemMetrics()
	if err != nil {
		log.Printf("failed to scrape system metrics: %v", err)
	}

	sysMetrics, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to gather metrics")
	}

	buf := bytes.NewBuffer([]byte{})

	for _, mf := range sysMetrics {
		_, err = expfmt.MetricFamilyToText(buf, mf)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expfmt to temporary buffer")
		}
	}

	sysMetricsSlice := strings.Split(buf.String(), "\n")
	allMetrics := slices.Concat(caddyMetrics, sysMetricsSlice)

	result := make([]string, 0, len(allMetrics))
	for _, s := range allMetrics {
		if strings.Contains(s, "ptah_") {
			if s[0] != '#' {
				s = fmt.Sprintf("%s %d", s, timestamp)
			}

			result = append(result, s)
		}
	}

	return result, nil
}

func (m *MetricsAgent) Start(ctx context.Context) error {
	log := Logger(ctx)

	log.Info("starting metrics agent")

	metrics, err := m.ScrapeMetrics(ctx)
	if err != nil {
		log.Info("failed to collect metrics", "error", err)
	}

	err = m.client.SendMetrics(ctx, metrics)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(m.interval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()

				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()

				metrics, err := m.ScrapeMetrics(ctx)
				if err != nil {
					log.Error("failed to collect metrics", "error", err)

					continue
				}

				err = m.client.SendMetrics(ctx, metrics)
				if err != nil {
					log.Error("failed to send metrics", "error", err)
				}
			}
		}
	}()

	return nil
}

func (m *MetricsAgent) Stop() {
	close(m.stopChan)
	m.wg.Wait()
}
