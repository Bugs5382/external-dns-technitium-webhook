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
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPassword = "pa55-do-not-log"
	testToken    = "tok-do-not-log"
)

// unreachableBase returns a base URL and port that refuse connections, so the
// HTTP client fails at the transport level and returns a *url.Error.
func unreachableBase(t *testing.T) (string, int) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())
	return "http://127.0.0.1", port
}

// A failed login must not echo the password back in the error, because
// external-dns logs provider errors verbatim (issue #28).
func TestLogin_TransportError_DoesNotLeakPassword(t *testing.T) {
	base, port := unreachableBase(t)
	c := NewClientWithCredentials(base, port, "admin", testPassword, true)

	err := c.Login()
	require.Error(t, err)
	assert.NotContains(t, err.Error(), testPassword)
	assert.Contains(t, err.Error(), "/api/user/login")
}

// A failed API call must not echo the session or API token in the error.
func TestDoRequest_TransportError_DoesNotLeakToken(t *testing.T) {
	base, port := unreachableBase(t)
	c := NewClientWithToken(base, port, testToken, true)

	_, err := c.DoRequest(http.MethodGet, "/api/zones/list", url.Values{"zone": {"example.test"}})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), testToken)
	assert.Contains(t, err.Error(), "/api/zones/list")
	assert.Contains(t, err.Error(), "example.test")
}

func TestRedactURL(t *testing.T) {
	got := redactURL("http://h:5380/api/user/login?user=admin&pass=" + testPassword + "&token=" + testToken)
	assert.NotContains(t, got, testPassword)
	assert.NotContains(t, got, testToken)
	assert.Contains(t, got, "user=admin")

	assert.Equal(t, "not a url %zz", redactURL("not a url %zz"))
}

// TECHNITIUM_SSL_VERIFY decides certificate verification; either way the
// client refuses anything older than TLS 1.2.
func TestCreateHTTPClient_TLSConfig(t *testing.T) {
	for _, verify := range []bool{true, false} {
		tr, ok := createHTTPClient(verify).Transport.(*http.Transport)
		require.True(t, ok)
		assert.Equal(t, !verify, tr.TLSClientConfig.InsecureSkipVerify)
		assert.Equal(t, uint16(tls.VersionTLS12), tr.TLSClientConfig.MinVersion)
	}
}
