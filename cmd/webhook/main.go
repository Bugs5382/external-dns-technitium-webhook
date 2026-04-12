// package main
package main

import (
	"fmt"

	"github.com/Bugs5382/external-dns-technitium-webhook/cmd/webhook/init/configuration"
	"github.com/Bugs5382/external-dns-technitium-webhook/cmd/webhook/init/dnsprovider"
	"github.com/Bugs5382/external-dns-technitium-webhook/cmd/webhook/init/logging"
	"github.com/Bugs5382/external-dns-technitium-webhook/cmd/webhook/init/server"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

const banner = `
  
external-dns-technitium-webhook
version: %s (%s)

`

var (
	// Version - value can be overridden by ldflags
	Version = "local"
	Gitsha  = "?"
)

func main() {
	fmt.Printf(banner, Version, Gitsha)

	// @todo remove
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	logging.Init()

	config := configuration.Init()
	provider, err := dnsprovider.Init(config)
	if err != nil {
		log.Fatalf("failed to initialize provider: %v", err)
	}

	srv := server.NewServer()

	srv.StartHealth(config)
	srv.Start(config, provider)
}
