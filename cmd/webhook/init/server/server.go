package server

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/Bugs5382/external-dns-technitium-webhook/cmd/webhook/init/configuration"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sigs.k8s.io/external-dns/provider"
	"sigs.k8s.io/external-dns/provider/webhook/api"
)

type WebhookServer struct {
	Ready   bool
	Channel chan struct{}
}

func NewServer() *WebhookServer {
	return &WebhookServer{
		Ready:   false,
		Channel: make(chan struct{}, 1),
	}
}

// Start Init server initialization function
// The server will respond to the following endpoints:
// - / (GET): initialization, negotiates headers and returns the domain filter
// - /records (GET): returns the current records
// - /records (POST): applies the changes
// - /adjustendpoints (POST): executes the AdjustEndpoints method
func (wh *WebhookServer) Start(config configuration.Config, p provider.Provider) {
	api.StartHTTPApi(p, wh.Channel, 0, 0, fmt.Sprintf("%s:%d", config.ServerHost, config.ServerPort))
}

func (wh *WebhookServer) StartHealth(config configuration.Config) {
	go func() {
		listenAddr := fmt.Sprintf("0.0.0.0:%d", config.HealthCheckPort)
		m := http.NewServeMux()
		m.Handle("/metrics", promhttp.Handler())
		m.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-wh.Channel:
				wh.Ready = true
			default:
			}
			if wh.Ready {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		})
		s := &http.Server{
			Addr:    listenAddr,
			Handler: m,
		}

		l, err := net.Listen("tcp", listenAddr)
		if err != nil {
			log.Fatal(err)
		}
		err = s.Serve(l)
		if err != nil {
			log.Fatalf("health listener stopped : %s", err)
		}
	}()
}
