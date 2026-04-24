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
	"testing"

	"github.com/Bugs5382/external-dns-technitium-webhook/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	cases := []struct {
		name          string
		cfg           config.Config
		env           map[string]string
		expectedError string
	}{
		{
			name: "minimal config for technitium provider (username/password)",
			cfg:  config.Config{},
			env: map[string]string{
				"TECHNITIUM_USER":     "user",
				"TECHNITIUM_PASSWORD": "password",
			},
		},
		{
			name: "minimal config for technitium provider (token)",
			cfg:  config.Config{},
			env: map[string]string{
				"TECHNITIUM_TOKEN": "token-123",
			},
		},
		{
			name: "domain filter config for technitium provider",
			cfg: config.Config{
				DomainFilter:   []string{"domain.com"},
				ExcludeDomains: []string{"sub.domain.com"},
			},
			env: map[string]string{
				"TECHNITIUM_USER":     "user",
				"TECHNITIUM_PASSWORD": "password",
			},
		},
		{
			name:          "empty configuration",
			cfg:           config.Config{},
			expectedError: "expecting error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			dnsProvider, err := Init(tc.cfg)

			if tc.expectedError != "" {
				assert.Error(t, err, "configuration error, no mandatory Environment variables set")
				return
			}

			assert.NoErrorf(t, err, "error creating provider")
			assert.NotNil(t, dnsProvider)
		})
	}
}
