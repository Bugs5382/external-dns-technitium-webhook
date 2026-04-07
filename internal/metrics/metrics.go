package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	FailedApiCallsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "external_dns_webhook_failed_api_calls_total",
		Help: "Total number of failed API calls",
	})
	TotalApiCalls = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "external_dns_webhook_total_api_calls",
		Help: "Total number of API calls",
	})

	ApiCallLatency = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "external_dns_webhook_api_call_latency_seconds",
		Help: "Latency of Infoblox API calls",
	}, []string{"method"}) // method label to differentiate API calls
)

func init() {
	prometheus.MustRegister(FailedApiCallsTotal, ApiCallLatency, TotalApiCalls)
}
