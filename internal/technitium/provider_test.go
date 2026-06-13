package technitium

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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/external-dns/endpoint"
)

// newTestProvider wires a Provider to a mock Technitium API that serves two
// zones (alpha.com, beta.org), each with a single A record host.<zone>.
func newTestProvider(t *testing.T, domainFilter *endpoint.DomainFilter) *Provider {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/zones/list":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"response": map[string]any{
					"zones": []map[string]string{{"name": "alpha.com"}, {"name": "beta.org"}},
				},
			})
		case "/api/zones/records/get":
			domain := r.URL.Query().Get("domain")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"response": map[string]any{
					"records": []map[string]any{
						{
							"name":  "host." + domain,
							"type":  "A",
							"ttl":   300,
							"rData": map[string]string{"ipAddress": "10.0.0.1"},
						},
					},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(ts.Close)

	u, err := url.Parse(ts.URL)
	require.NoError(t, err)

	return &Provider{
		client: &Client{
			BaseURL:       u.Scheme + "://" + u.Hostname(),
			Port:          u.Port(),
			token:         "test-token",
			isStaticToken: true,
			HTTPClient:    ts.Client(),
		},
		domainFilter: domainFilter,
		config:       &StartupConfig{},
	}
}

func recordNames(endpoints []*endpoint.Endpoint) []string {
	names := make([]string, 0, len(endpoints))
	for _, ep := range endpoints {
		names = append(names, ep.DNSName)
	}
	return names
}

// With no domain filter configured every zone is in scope, so all FQDNs are
// returned.
func TestRecords_NoFilter_ReturnsAllFQDNs(t *testing.T) {
	p := newTestProvider(t, endpoint.NewDomainFilterWithExclusions(nil, nil))

	endpoints, err := p.Records(context.Background())
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"host.alpha.com", "host.beta.org"}, recordNames(endpoints))
}

// A nil filter must also be treated as match-all rather than skipping every
// zone (guards against a panic / empty result).
func TestRecords_NilFilter_ReturnsAllFQDNs(t *testing.T) {
	p := newTestProvider(t, nil)

	endpoints, err := p.Records(context.Background())
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"host.alpha.com", "host.beta.org"}, recordNames(endpoints))
}

// A configured filter scopes discovery to the matching zones only.
func TestRecords_WithFilter_ScopesToMatchingZones(t *testing.T) {
	p := newTestProvider(t, endpoint.NewDomainFilterWithExclusions([]string{"alpha.com"}, nil))

	endpoints, err := p.Records(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []string{"host.alpha.com"}, recordNames(endpoints))
}
