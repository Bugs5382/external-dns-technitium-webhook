package server

/*
Apache License 2.0

Copyright 2026 external-dns-technitium-webhook Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"fmt"
	"net"
	"net/http"

	"github.com/Bugs5382/external-dns-technitium-webhook/internal/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
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

func (wh *WebhookServer) Start(cfg config.Config, p provider.Provider) {
	api.StartHTTPApi(p, wh.Channel, 0, 0, fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort))
}

func (wh *WebhookServer) StartHealth(cfg config.Config) {
	go func() {
		listenAddr := fmt.Sprintf("0.0.0.0:%d", cfg.HealthCheckPort)
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
			log.Fatal().Msgf("%s", err)
		}
		if err = s.Serve(l); err != nil {
			log.Fatal().Msgf("health listener stopped: %s", err)
		}
	}()
}
