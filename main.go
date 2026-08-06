package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/vinted/postfix-exporter/collector"
)

var (
	bindAddr         = flag.String("telemetry.addr", ":9706", "host:port for postfix exporter")
	queryInterval    = flag.Int("query.interval", 15, "How often should daemon read metrics")
	logLevel         = flag.String("log.level", "info", "Logging level")
	postfixSpoolPath = flag.String("spool.path", "/var/spool/postfix", "path to Postfix spool directory")
	postfixLogPath   = flag.String("log.path", "/var/log/maillog", "path to Postfix log file for delivery metrics, empty to disable")
)

func main() {

	flag.Parse()

	switch *logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	go collector.CollectTimer(*queryInterval, *postfixSpoolPath)

	pf := collector.NewPostfixCollector()
	prometheus.MustRegister(pf)

	if *postfixLogPath != "" {
		prometheus.MustRegister(collector.DeliveriesTotal, collector.DeliveryDelay,
			collector.DeliveryStageDelay, collector.LogTailActive)
		go collector.TailLog(*postfixLogPath)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
    <head><title>Postfix Exporter</title></head>
    <body>
    <h1>Postfix Exporter</h1>
    <p><a href='metrics'>Metrics</a></p>
    </body>
    </html>`))
		if err != nil {
			log.Error("HTTP write failed: ", err)
		}
	})
	log.Info("Listening on: ", *bindAddr)
	log.Fatal(http.ListenAndServe(*bindAddr, nil))
}
