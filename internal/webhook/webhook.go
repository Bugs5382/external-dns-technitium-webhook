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

	"github.com/Bugs5382/external-dns-technitium-webhook/internal/config"
	"github.com/Bugs5382/external-dns-technitium-webhook/internal/technitium"
	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog/log"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"
)

func Init(cfg config.Config) (provider.Provider, error) {
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
