// package main
package main

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

	"github.com/Bugs5382/external-dns-technitium-webhook/cmd/webhook/init/configuration"
	"github.com/Bugs5382/external-dns-technitium-webhook/cmd/webhook/init/dnsprovider"
	"github.com/Bugs5382/external-dns-technitium-webhook/cmd/webhook/init/logging"
	"github.com/Bugs5382/external-dns-technitium-webhook/cmd/webhook/init/server"
	"github.com/rs/zerolog/log"
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

	logging.Init()

	config := configuration.Init()
	provider, err := dnsprovider.Init(config)
	if err != nil {
		log.Fatal().Msgf("failed to initialize provider: %v", err)
	}

	srv := server.NewServer()

	srv.StartHealth(config)
	srv.Start(config, provider)
}
