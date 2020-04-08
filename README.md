# Prometheus Postfix exporter

**Information**

Exporter collects Postfix queue metrics:

`maildrop`   
`hold`   
`incoming`   
`active`   
`defer`   

Uses separate collection thread, so scrape time will not be affected on highly loaded mail servers.


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
  -spool.path sting
      path to Postfix spool directory (default "/var/spool/postfix")
```
