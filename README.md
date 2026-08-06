# Prometheus Postfix exporter

**Information**

Exporter collects Postfix queue metrics:

`maildrop`
`hold`
`incoming`
`active`
`deferred`

Uses separate collection thread, so scrape time will not be affected on highly loaded mail servers.

Exporter also tails the Postfix log file and collects delivery metrics:

`postfix_deliveries_total{status}` — delivery attempts by status (`sent`, `deferred`, `bounced`, ...)
`postfix_delivery_delay_seconds` — total delay histogram of sent messages
`postfix_delivery_stage_delay_seconds{stage}` — per-stage delay histogram of sent messages, `stage` is one of
the Postfix `delays=` stages (`before_queue_manager`, `queue_manager`, `connection_setup`, `transmission`)
`postfix_log_tail_active` — 1 while the log tail is running, 0 while it is stopped and waiting to retry

Log tailing starts at the end of the file (restarts do not replay old lines) and survives log rotation.
The exporter process must have read access to the log file; if it does not, an error is logged,
`postfix_log_tail_active` drops to 0 and the tail is retried every 30 seconds.
`postfix_deliveries_total` counts per-recipient delivery attempts, so a
message retried or sent to several recipients increments it more than once.


**Building**

Checkout https://github.com/vinted/postfix-exporter repo.  
Build executable:  

 `go build`

**Using**

Execute postfix-exporter:  

`./postfix-exporter`

By default exporter will bind to port `9706`.  

**Configuration**

Following config parameters are available:  

```
  -telemetry.addr string
    	host:port for postfix exporter (default ":9706")
  -query.interval int
      How often should daemon read metric (default 15)
  -log.level string
      Logging level (default "info")
  -spool.path string
      path to Postfix spool directory (default "/var/spool/postfix")
  -log.path string
      path to Postfix log file for delivery metrics, empty to disable (default "/var/log/maillog")
```
