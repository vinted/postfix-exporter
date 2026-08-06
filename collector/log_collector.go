package collector

import (
	"io"
	"regexp"
	"strconv"
	"time"

	"github.com/nxadm/tail"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Postfix logs delay values with %g, so large values use scientific notation (1.2e+05).
const delayNumber = `[0-9.]+(?:e[+-]?[0-9]+)?`

// Matches delivery agent log lines, e.g.:
// postfix/smtp[1234]: 4XYZ12abc: to=<user@example.com>, relay=..., delay=1.4,
// delays=0.02/0.01/0.5/0.9, dsn=2.0.0, status=sent (250 2.0.0 OK)
var deliveryRegexp = regexp.MustCompile(
	`postfix(?:-[\w.]+)?/(?:smtp|lmtp|local|virtual|pipe|relay|error|discard)\[\d+\]: ` +
		`[0-9A-Za-z]+: to=<[^>]*>, .*delay=(` + delayNumber + `), ` +
		`delays=(` + delayNumber + `)/(` + delayNumber + `)/(` + delayNumber + `)/(` + delayNumber + `), ` +
		`.*status=(\w+)`)

var delayStages = [...]string{"before_queue_manager", "queue_manager", "connection_setup", "transmission"}

var delayBuckets = []float64{0.1, 0.5, 1, 2.5, 5, 10, 30, 60, 300, 900, 3600, 14400, 86400}

const tailRetryInterval = 30 * time.Second

var (
	DeliveriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "postfix_deliveries_total",
			Help: "Number of delivery attempts logged by Postfix delivery agents, by status",
		},
		[]string{"status"},
	)
	DeliveryDelay = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "postfix_delivery_delay_seconds",
			Help:    "Total delay of successfully delivered messages",
			Buckets: delayBuckets,
		},
	)
	DeliveryStageDelay = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "postfix_delivery_stage_delay_seconds",
			Help:    "Delay of successfully delivered messages per Postfix delays= stage",
			Buckets: delayBuckets,
		},
		[]string{"stage"},
	)
	LogTailActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "postfix_log_tail_active",
			Help: "Whether the Postfix log tail is running (1) or stopped and waiting to retry (0)",
		},
	)
)

type delivery struct {
	status string
	delay  float64
	delays [len(delayStages)]float64
}

func parseDeliveryLine(line string) (delivery, bool) {
	match := deliveryRegexp.FindStringSubmatch(line)
	if match == nil {
		return delivery{}, false
	}
	var d delivery
	var err error
	d.delay, err = strconv.ParseFloat(match[1], 64)
	if err != nil {
		return delivery{}, false
	}
	for i := range d.delays {
		d.delays[i], err = strconv.ParseFloat(match[2+i], 64)
		if err != nil {
			return delivery{}, false
		}
	}
	d.status = match[len(delayStages)+2]
	return d, true
}

func (d delivery) observe() {
	DeliveriesTotal.WithLabelValues(d.status).Inc()
	if d.status != "sent" {
		return
	}
	DeliveryDelay.Observe(d.delay)
	for i, stage := range delayStages {
		DeliveryStageDelay.WithLabelValues(stage).Observe(d.delays[i])
	}
}

func TailLog(logPath string) {
	for {
		tailLog(logPath)
		LogTailActive.Set(0)
		time.Sleep(tailRetryInterval)
	}
}

func tailLog(logPath string) {
	t, err := tail.TailFile(logPath, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: false,
		Location:  &tail.SeekInfo{Offset: 0, Whence: io.SeekEnd},
		Logger:    log.StandardLogger(),
	})
	if err != nil {
		log.Error("log tail failed to start: ", err)
		return
	}
	defer t.Cleanup()
	LogTailActive.Set(1)
	for line := range t.Lines {
		if line.Err != nil {
			log.Debug(line.Err)
			continue
		}
		if d, ok := parseDeliveryLine(line.Text); ok {
			d.observe()
		}
	}
	log.Error("log tail stopped, retrying: ", t.Err())
}
