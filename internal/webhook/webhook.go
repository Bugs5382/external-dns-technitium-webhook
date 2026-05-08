package webhook

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
	"regexp"
	"strings"
	"time"

	"github.com/Bugs5382/external-dns-technitium-webhook/internal/config"
	"github.com/Bugs5382/external-dns-technitium-webhook/internal/technitium"
	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog/log"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"
)

// pauseInterval controls how often Pause re-emits the configuration
// error while the container is held alive. Exposed as a var so tests
// can shorten it.
var pauseInterval = 5 * time.Minute

// Pause holds the goroutine forever after logging the configuration
// error that prevented startup. It is the failure mode for invalid
// configuration so the container stays Running with visible logs
// instead of looping through CrashLoopBackOff and losing the message.
func Pause(err error) {
	log.Error().Err(err).Msg("configuration error — container paused, fix the environment and restart the pod")
	for {
		time.Sleep(pauseInterval)
		log.Warn().Err(err).Msg("still paused — fix the environment and restart the pod")
	}
}

func hasDomainScope(cfg config.Config) bool {
	for _, d := range cfg.DomainFilter {
		if d != "" {
			return true
		}
	}
	for _, d := range cfg.ExcludeDomains {
		if d != "" {
			return true
		}
	}
	return cfg.RegexDomainFilter != "" ||
		cfg.RegexDomainExclusion != "" ||
		cfg.RegexNameFilter != ""
}

func Init(cfg config.Config) (provider.Provider, error) {
	if !hasDomainScope(cfg) {
		return nil, fmt.Errorf("no domain scope configured: set at least one of DOMAIN_FILTER, EXCLUDE_DOMAIN_FILTER, REGEXP_DOMAIN_FILTER, REGEXP_DOMAIN_FILTER_EXCLUSION, REGEXP_NAME_FILTER — running without one would let this webhook claim every domain")
	}

	var domainFilter *endpoint.DomainFilter
	createMsg := "Creating technitium provider with "

	if cfg.RegexDomainFilter != "" {
		createMsg += fmt.Sprintf("regexp domain filter: '%s', ", cfg.RegexDomainFilter)
		if cfg.RegexDomainExclusion != "" {
			createMsg += fmt.Sprintf("with exclusion: '%s', ", cfg.RegexDomainExclusion)
		}
		domainFilter = endpoint.NewRegexDomainFilter(
			regexp.MustCompile(cfg.RegexDomainFilter),
			regexp.MustCompile(cfg.RegexDomainExclusion),
		)
	} else {
		if len(cfg.DomainFilter) > 0 {
			createMsg += fmt.Sprintf("domain filter: '%s', ", strings.Join(cfg.DomainFilter, ","))
		}
		if len(cfg.ExcludeDomains) > 0 {
			createMsg += fmt.Sprintf("exclude domain filter: '%s', ", strings.Join(cfg.ExcludeDomains, ","))
		}
		domainFilter = endpoint.NewDomainFilterWithExclusions(cfg.DomainFilter, cfg.ExcludeDomains)
	}

	createMsg = strings.TrimSuffix(createMsg, ", ")
	if strings.HasSuffix(createMsg, "with ") {
		createMsg += "no kind of domain filters"
	}

	log.Info().Msg(createMsg)

	technitiumConfig := technitium.StartupConfig{}
	if err := env.Parse(&technitiumConfig); err != nil {
		return nil, fmt.Errorf("reading configuration failed: %v", err)
	}

	hasToken := technitiumConfig.Token != ""
	hasCredentials := technitiumConfig.Username != "" && technitiumConfig.Password != ""

	if !hasToken && !hasCredentials {
		return nil, fmt.Errorf("missing credentials: you must provide either TECHNITIUM_TOKEN, or both TECHNITIUM_USER and TECHNITIUM_PASSWORD")
	}

	technitiumConfig.FQDNRegEx = cfg.RegexDomainFilter
	technitiumConfig.NameRegEx = cfg.RegexNameFilter

	if hasToken {
		return technitium.NewTechnitiumProviderWithToken(&technitiumConfig, domainFilter)
	}

	return technitium.NewTechnitiumProviderWithCredentials(&technitiumConfig, domainFilter)
}
