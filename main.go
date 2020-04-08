package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/vinted/postfix-exporter/collector"
	"net/http"
)

var (
	bindAddr         = flag.String("telemetry.addr", ":9706", "host:port for postfix exporter")
	queryInterval    = flag.Int("query.interval", 15, "How often should daemon read metrics")
	logLevel         = flag.String("log.level", "info", "Logging level")
	postfixSpoolPath = flag.String("spool.path", "/var/spool/postfix", "path to Postfix spool directory")
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

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
    <head><title>Ceph Exporter</title></head>
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
