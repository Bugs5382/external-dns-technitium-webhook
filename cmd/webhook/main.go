// package main
package main

import (
	"fmt"

	"github.com/Bugs5382/external-dns-technitium-webhook/cmd/webhook/logging"
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
}
