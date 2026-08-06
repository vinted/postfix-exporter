package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestParseDeliveryLine(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		want   delivery
		wantOk bool
	}{
		{
			name: "sent",
			line: "Aug  5 10:00:00 mail1 postfix/smtp[1234]: 4XYZ12abc: to=<user@example.com>, " +
				"relay=gmail-smtp-in.l.google.com[142.250.27.26]:25, delay=1.4, delays=0.02/0.01/0.5/0.9, " +
				"dsn=2.0.0, status=sent (250 2.0.0 OK)",
			want:   delivery{status: "sent", delay: 1.4, delays: [4]float64{0.02, 0.01, 0.5, 0.9}},
			wantOk: true,
		},
		{
			name: "deferred with conn_use",
			line: "Aug  5 10:00:00 mail1 postfix/smtp[1234]: A1B2C3: to=<user@example.com>, " +
				"relay=mx.example.com[10.0.0.1]:25, conn_use=2, delay=7207, delays=7200/1/2/4, " +
				"dsn=4.4.2, status=deferred (lost connection)",
			want:   delivery{status: "deferred", delay: 7207, delays: [4]float64{7200, 1, 2, 4}},
			wantOk: true,
		},
		{
			name: "bounced with orig_to",
			line: "Aug  5 10:00:00 relay1 postfix/relay[99]: B2C3D4: to=<a@b.fr>, orig_to=<root>, " +
				"relay=mailgw.example.net[192.0.2.25]:25, delay=0.5, delays=0.1/0/0.2/0.2, " +
				"dsn=5.1.1, status=bounced (user unknown)",
			want:   delivery{status: "bounced", delay: 0.5, delays: [4]float64{0.1, 0, 0.2, 0.2}},
			wantOk: true,
		},
		{
			name: "local delivery",
			line: "Aug  5 10:00:00 host postfix/local[7]: C3D4E5: to=<root@host>, relay=local, " +
				"delay=0.02, delays=0.01/0/0/0.01, dsn=2.0.0, status=sent (delivered to mailbox)",
			want:   delivery{status: "sent", delay: 0.02, delays: [4]float64{0.01, 0, 0, 0.01}},
			wantOk: true,
		},
		{
			name: "deferred with scientific notation delays",
			line: "Aug  5 10:00:00 mail1 postfix/smtp[1234]: D4E5F6: to=<user@example.com>, " +
				"relay=none, delay=1.2e+05, delays=1.2e+05/0.02/1.1/2.3, " +
				"dsn=4.4.1, status=deferred (connect to mx.example.com[10.0.0.1]:25: Connection timed out)",
			want:   delivery{status: "deferred", delay: 1.2e+05, delays: [4]float64{1.2e+05, 0.02, 1.1, 2.3}},
			wantOk: true,
		},
		{
			name:   "smtpd connect is ignored",
			line:   "Aug  5 10:00:00 host postfix/smtpd[5]: connect from client.example.com[10.0.0.2]",
			wantOk: false,
		},
		{
			name:   "qmgr removed is ignored",
			line:   "Aug  5 10:00:00 host postfix/qmgr[6]: 4XYZ12abc: removed",
			wantOk: false,
		},
		{
			name:   "qmgr expired is ignored",
			line:   "Aug  5 10:00:00 host postfix/qmgr[6]: E5F6A7: from=<sender@example.com>, status=expired, returned to sender",
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseDeliveryLine(tt.line)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	deferred := delivery{status: "deferred", delay: 7207, delays: [4]float64{7200, 1, 2, 4}}
	deferred.observe()
	if got := testutil.ToFloat64(DeliveriesTotal.WithLabelValues("deferred")); got != 1 {
		t.Errorf("deliveries_total{status=deferred} = %v, want 1", got)
	}
	if got := testutil.CollectAndCount(DeliveryStageDelay); got != 0 {
		t.Errorf("stage delay series after deferred = %d, want 0", got)
	}

	sent := delivery{status: "sent", delay: 1.4, delays: [4]float64{0.02, 0.01, 0.5, 0.9}}
	sent.observe()
	if got := testutil.ToFloat64(DeliveriesTotal.WithLabelValues("sent")); got != 1 {
		t.Errorf("deliveries_total{status=sent} = %v, want 1", got)
	}
	if got := testutil.CollectAndCount(DeliveryStageDelay); got != len(delayStages) {
		t.Errorf("stage delay series after sent = %d, want %d", got, len(delayStages))
	}
}
