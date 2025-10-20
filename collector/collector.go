package collector

import (
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type postfixCollector struct {
	maildropCount  *prometheus.Desc
	maildropSize   *prometheus.Desc
	holdCount      *prometheus.Desc
	holdSize       *prometheus.Desc
	incomingCount  *prometheus.Desc
	incomingSize   *prometheus.Desc
	activeCount    *prometheus.Desc
	activeSize     *prometheus.Desc
	deferCount     *prometheus.Desc
	deferSize      *prometheus.Desc
	deferredCount  *prometheus.Desc
	deferredSize   *prometheus.Desc
	collectionTime *prometheus.Desc
	scrapeTime     *prometheus.Desc
}

type postfixMetrics struct {
	maildropCount  float64
	maildropSize   float64
	holdCount      float64
	holdSize       float64
	incomingCount  float64
	incomingSize   float64
	activeCount    float64
	activeSize     float64
	deferCount     float64
	deferSize      float64
	deferredCount  float64
	deferredSize   float64
	collectionTime float64
}

var metricsCtx postfixMetrics

func NewPostfixCollector() *postfixCollector {
	return &postfixCollector{
		maildropCount: prometheus.NewDesc("postfix_maildrop_queue_count",
			"Number of messages in maildrop queue",
			nil, nil,
		),
		maildropSize: prometheus.NewDesc("postfix_maildrop_queue_size",
			"Total size of messages in maildrop queue",
			nil, nil,
		),
		holdCount: prometheus.NewDesc("postfix_hold_queue_count",
			"Number of messages in hold queue",
			nil, nil,
		),
		holdSize: prometheus.NewDesc("postfix_hold_queue_size",
			"Total size of messages in hold queue",
			nil, nil,
		),
		incomingCount: prometheus.NewDesc("postfix_incoming_queue_count",
			"Number of messages in incoming queue",
			nil, nil,
		),
		incomingSize: prometheus.NewDesc("postfix_incoming_queue_size",
			"Total size of messages in incoming queue",
			nil, nil,
		),
		activeCount: prometheus.NewDesc("postfix_active_queue_count",
			"Number of messages in active queue",
			nil, nil,
		),
		activeSize: prometheus.NewDesc("postfix_active_queue_size",
			"Total size of messages in active queue",
			nil, nil,
		),
		deferCount: prometheus.NewDesc("postfix_defer_queue_count",
			"Number of messages in defer queue",
			nil, nil,
		),
		deferSize: prometheus.NewDesc("postfix_defer_queue_size",
			"Total size of messages in defer queue",
			nil, nil,
		),
		deferredCount: prometheus.NewDesc("postfix_deferred_queue_count",
			"Number of messages in deferred queue",
			nil, nil,
		),
		deferredSize: prometheus.NewDesc("postfix_deferred_queue_size",
			"Total size of messages in deferred queue",
			nil, nil,
		),
		collectionTime: prometheus.NewDesc("postfix_metric_collection_time",
			"Time it took for a collection thread to collect postfix metrics",
			nil, nil,
		),
		scrapeTime: prometheus.NewDesc("postfix_metric_scrape_time",
			"Time it took for prometheus to scrape postfix metrics",
			nil, nil,
		),
	}
}

func (collector *postfixCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.maildropCount
	ch <- collector.maildropSize
	ch <- collector.holdCount
	ch <- collector.holdSize
	ch <- collector.incomingCount
	ch <- collector.incomingSize
	ch <- collector.activeCount
	ch <- collector.activeSize
	ch <- collector.deferCount
	ch <- collector.deferSize
	ch <- collector.deferredCount
	ch <- collector.deferredSize
	ch <- collector.collectionTime
	ch <- collector.scrapeTime
}

func (collector *postfixCollector) Collect(ch chan<- prometheus.Metric) {
	scrapeTime := time.Now()
	ch <- prometheus.MustNewConstMetric(collector.maildropCount, prometheus.GaugeValue, metricsCtx.maildropCount)
	ch <- prometheus.MustNewConstMetric(collector.maildropSize, prometheus.GaugeValue, metricsCtx.maildropSize)
	ch <- prometheus.MustNewConstMetric(collector.holdCount, prometheus.GaugeValue, metricsCtx.holdCount)
	ch <- prometheus.MustNewConstMetric(collector.holdSize, prometheus.GaugeValue, metricsCtx.holdSize)
	ch <- prometheus.MustNewConstMetric(collector.incomingCount, prometheus.GaugeValue, metricsCtx.incomingCount)
	ch <- prometheus.MustNewConstMetric(collector.incomingSize, prometheus.GaugeValue, metricsCtx.incomingSize)
	ch <- prometheus.MustNewConstMetric(collector.activeCount, prometheus.GaugeValue, metricsCtx.activeCount)
	ch <- prometheus.MustNewConstMetric(collector.activeSize, prometheus.GaugeValue, metricsCtx.activeSize)
	ch <- prometheus.MustNewConstMetric(collector.deferCount, prometheus.GaugeValue, metricsCtx.deferCount)
	ch <- prometheus.MustNewConstMetric(collector.deferSize, prometheus.GaugeValue, metricsCtx.deferSize)
	ch <- prometheus.MustNewConstMetric(collector.deferredCount, prometheus.GaugeValue, metricsCtx.deferredCount)
	ch <- prometheus.MustNewConstMetric(collector.deferredSize, prometheus.GaugeValue, metricsCtx.deferredSize)
	ch <- prometheus.MustNewConstMetric(collector.collectionTime, prometheus.GaugeValue, metricsCtx.collectionTime)
	ch <- prometheus.MustNewConstMetric(collector.scrapeTime, prometheus.GaugeValue, time.Since(scrapeTime).Seconds())
}

func CollectTimer(queryInterval int, postfixSpoolPath string) {
	tickChan := time.NewTicker(time.Second * time.Duration(queryInterval))
	defer tickChan.Stop()
	for range tickChan.C {
		log.Debug("collectMetrics triggered")
		metricsCtx.collectMetrics(postfixSpoolPath)
		log.Debug("collectMetrics ended")
	}
}

func (pm *postfixMetrics) collectMetrics(postfixSpoolPath string) {
	collectionTime := time.Now()
	pm.maildropCount, pm.maildropSize = DirectoryWalk(postfixSpoolPath, "maildrop")
	pm.holdCount, pm.holdSize = DirectoryWalk(postfixSpoolPath, "hold")
	pm.incomingCount, pm.incomingSize = DirectoryWalk(postfixSpoolPath, "incoming")
	pm.activeCount, pm.activeSize = DirectoryWalk(postfixSpoolPath, "active")
	pm.deferCount, pm.deferSize = DirectoryWalk(postfixSpoolPath, "defer")
	pm.deferredCount, pm.deferredSize = DirectoryWalk(postfixSpoolPath, "deferred")
	pm.collectionTime = time.Since(collectionTime).Seconds()
}

func DirectoryWalk(postfixSpoolPath, dirName string) (float64, float64) {
	var count, size float64
	directory := postfixSpoolPath + "/" + dirName
	err := filepath.Walk(directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Debug(err)
			} else if !info.IsDir() {
				size = size + float64(info.Size())
				count++
			}
			return nil
		},
	)
	if err != nil {
		log.Error(err)
	}
	return count, size
}
