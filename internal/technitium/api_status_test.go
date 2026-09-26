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
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

// Technitium answers HTTP 200 for failed calls and puts the outcome in the
// JSON "status" field. These bodies are copied from Technitium 15.5.1
// (issue #29).
const (
	bodyOK           = `{"server":"t","status":"ok","response":{}}`
	bodyNoSuchZone   = `{"server":"t","status":"error","errorMessage":"No such zone was found: x.nozone.invalid"}`
	bodyInvalidToken = `{"server":"t","status":"invalid-token","errorMessage":"Invalid token or session expired."}`
	bodyLoginOK      = `{"server":"t","status":"ok","token":"session-token"}`
)

type mockAPI struct {
	logins atomic.Int32
	// apiBody is served for every call except user/login.
	apiBody string
}

func newMockAPI(t *testing.T, apiBody string) (*mockAPI, string, int) {
	t.Helper()
	m := &mockAPI{apiBody: apiBody}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/user/login" {
			m.logins.Add(1)
			_, _ = w.Write([]byte(bodyLoginOK))
			return
		}
		_, _ = w.Write([]byte(m.apiBody))
	}))
	t.Cleanup(ts.Close)

	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	port, err := strconv.Atoi(u.Port())
	require.NoError(t, err)
	return m, u.Scheme + "://" + u.Hostname(), port
}

func TestDoRequest_StatusOK_ReturnsBody(t *testing.T) {
	_, base, port := newMockAPI(t, bodyOK)
	c := NewClientWithToken(base, port, testToken, true)

	body, err := c.DoRequest(http.MethodGet, "/api/zones/list", nil)
	require.NoError(t, err)
	assert.JSONEq(t, bodyOK, string(body))
}

func TestDoRequest_StatusError_ReturnsAPIError(t *testing.T) {
	_, base, port := newMockAPI(t, bodyNoSuchZone)
	c := NewClientWithToken(base, port, testToken, true)

	_, err := c.DoRequest(http.MethodGet, "/api/zones/records/add", url.Values{"domain": {"x.nozone.invalid"}})
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "/api/zones/records/add", apiErr.Path)
	assert.Equal(t, "error", apiErr.Status)
	assert.Equal(t, "No such zone was found: x.nozone.invalid", apiErr.Message)
	assert.NotErrorIs(t, err, ErrInvalidToken)
	assert.NotContains(t, err.Error(), testToken)
}

// A body that is not the Technitium JSON envelope is a failure too, rather
// than being handed to the caller as if it were data.
func TestDoRequest_NonJSONBody_ReturnsError(t *testing.T) {
	_, base, port := newMockAPI(t, "<html>proxy error</html>")
	c := NewClientWithToken(base, port, testToken, true)

	_, err := c.DoRequest(http.MethodGet, "/api/zones/list", nil)
	require.Error(t, err)
}

// A session client must drop a token the server has rejected, so the next
// call logs in again instead of reusing it.
func TestDoRequest_InvalidToken_SessionClientLogsInAgain(t *testing.T) {
	m, base, port := newMockAPI(t, bodyInvalidToken)
	c := NewClientWithCredentials(base, port, "admin", testPassword, true)
	require.NoError(t, c.Login())
	c.tokenExpiry = time.Now().Add(time.Hour) // keep the session "fresh" so only the rejection can clear it

	_, err := c.DoRequest(http.MethodGet, "/api/zones/list", nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidToken)
	assert.Empty(t, c.token, "rejected session token must be cleared")

	m.apiBody = bodyOK
	_, err = c.DoRequest(http.MethodGet, "/api/zones/list", nil)
	require.NoError(t, err)
	assert.Equal(t, int32(2), m.logins.Load(), "second call must log in again")
}

// A static token cannot be refreshed; the rejection is returned and the
// token is kept (there is nothing to replace it with).
func TestDoRequest_InvalidToken_StaticToken(t *testing.T) {
	m, base, port := newMockAPI(t, bodyInvalidToken)
	c := NewClientWithToken(base, port, testToken, true)

	_, err := c.DoRequest(http.MethodGet, "/api/zones/list", nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidToken)
	assert.NotContains(t, err.Error(), testToken)
	assert.Equal(t, testToken, c.token)
	assert.Equal(t, int32(0), m.logins.Load())
}

func newStatusProvider(t *testing.T, apiBody string) *Provider {
	t.Helper()
	_, base, port := newMockAPI(t, apiBody)
	return &Provider{
		client:       NewClientWithToken(base, port, testToken, true),
		domainFilter: endpoint.NewDomainFilter([]string{"example.test"}),
		config:       &StartupConfig{},
	}
}

// A rejected write must reach external-dns as an error, so the change is not
// recorded as applied and is retried on the next sync.
func TestApplyChanges_RejectedWrite_ReturnsError(t *testing.T) {
	p := newStatusProvider(t, bodyNoSuchZone)

	err := p.ApplyChanges(context.Background(), &plan.Changes{
		Create: []*endpoint.Endpoint{endpoint.NewEndpoint("x.nozone.invalid", "A", "10.0.0.1")},
	})
	require.Error(t, err)
	var apiErr *APIError
	assert.True(t, errors.As(err, &apiErr))
}

// One rejected record must not block the rest of the batch: every change is
// attempted and the failures are returned together.
func TestApplyChanges_RejectedWrite_OtherChangesStillApplied(t *testing.T) {
	var added []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domain := r.URL.Query().Get("domain")
		if domain == "x.nozone.invalid" {
			_, _ = w.Write([]byte(bodyNoSuchZone))
			return
		}
		added = append(added, domain)
		_, _ = w.Write([]byte(bodyOK))
	}))
	t.Cleanup(ts.Close)
	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	port, err := strconv.Atoi(u.Port())
	require.NoError(t, err)

	p := &Provider{
		client: NewClientWithToken(u.Scheme+"://"+u.Hostname(), port, testToken, true),
		config: &StartupConfig{},
	}
	err = p.ApplyChanges(context.Background(), &plan.Changes{
		Create: []*endpoint.Endpoint{
			endpoint.NewEndpoint("x.nozone.invalid", "A", "10.0.0.1"),
			endpoint.NewEndpoint("ok.example.test", "A", "10.0.0.2"),
		},
		Delete: []*endpoint.Endpoint{endpoint.NewEndpoint("old.example.test", "A", "10.0.0.3")},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x.nozone.invalid")
	assert.Equal(t, []string{"ok.example.test", "old.example.test"}, added)
}

// An error body from zones/list must not be read as "no zones", which would
// make external-dns think every record is gone.
func TestRecords_ZoneListError_ReturnsError(t *testing.T) {
	p := newStatusProvider(t, bodyInvalidToken)

	eps, err := p.Records(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidToken)
	assert.Nil(t, eps)
}
