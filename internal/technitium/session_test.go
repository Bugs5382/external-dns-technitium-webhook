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
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A username and password client logs in once and reuses the session for
// later calls (issue #31).
func TestDoRequest_ReusesSession(t *testing.T) {
	m, base, port := newMockAPI(t, bodyOK)
	c := NewClientWithCredentials(base, port, "admin", testPassword, true)

	for range 5 {
		_, err := c.DoRequest(http.MethodGet, "/api/zones/list", nil)
		require.NoError(t, err)
	}
	assert.Equal(t, int32(1), m.logins.Load())
}

// Once the session lifetime has passed, the next call logs in again.
func TestDoRequest_ExpiredSession_LogsInAgain(t *testing.T) {
	m, base, port := newMockAPI(t, bodyOK)
	c := NewClientWithCredentials(base, port, "admin", testPassword, true)

	_, err := c.DoRequest(http.MethodGet, "/api/zones/list", nil)
	require.NoError(t, err)
	c.tokenExpiry = time.Now().Add(-time.Second)

	_, err = c.DoRequest(http.MethodGet, "/api/zones/list", nil)
	require.NoError(t, err)
	assert.Equal(t, int32(2), m.logins.Load())
}

// TECHNITIUM_SESSION_TTL=0 matches a Technitium session timeout of 0: the
// session never expires on the client side and is replaced only when the
// server rejects it.
func TestDoRequest_ZeroTTL_NeverExpiresLocally(t *testing.T) {
	m, base, port := newMockAPI(t, bodyOK)
	c := NewClientWithCredentials(base, port, "admin", testPassword, true)
	c.SessionTTL = 0

	for range 3 {
		_, err := c.DoRequest(http.MethodGet, "/api/zones/list", nil)
		require.NoError(t, err)
	}
	assert.Equal(t, int32(1), m.logins.Load())
	assert.True(t, c.tokenExpiry.IsZero())
}

func TestSessionLifetime(t *testing.T) {
	assert.Equal(t, 27*time.Minute, sessionLifetime(30*time.Minute))
	assert.Equal(t, 54*time.Second, sessionLifetime(time.Minute))
	assert.Equal(t, time.Duration(0), sessionLifetime(0))
	assert.Equal(t, time.Duration(0), sessionLifetime(-time.Minute))
}

func TestNewClientWithCredentials_DefaultSessionTTL(t *testing.T) {
	c := NewClientWithCredentials("http://h", 5380, "admin", testPassword, true)
	assert.Equal(t, defaultSessionTTL, c.SessionTTL)
}

// The provider hands TECHNITIUM_SESSION_TTL (minutes) to the client.
func TestNewTechnitiumProviderWithCredentials_UsesSessionTTL(t *testing.T) {
	p, err := NewTechnitiumProviderWithCredentials(&StartupConfig{
		Host: "http://h", Port: 5380, Username: "admin", Password: testPassword, SessionTTL: 5,
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, 5*time.Minute, p.client.SessionTTL)
}

// The TLS-skip warning only means something for an https host.
func TestWarnInsecureTLS(t *testing.T) {
	assert.True(t, warnInsecureTLS("https://technitium", false))
	assert.True(t, warnInsecureTLS("HTTPS://technitium", false))
	assert.False(t, warnInsecureTLS("https://technitium", true))
	assert.False(t, warnInsecureTLS("http://technitium", false))
	assert.False(t, warnInsecureTLS("technitium", false))
}
