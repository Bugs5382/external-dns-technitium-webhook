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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Bugs5382/external-dns-technitium-webhook/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

func NewClientWithCredentials(baseURL string, port int, username, password string, sslVerify bool) *Client {
	return &Client{
		BaseURL:    baseURL,
		Port:       strconv.Itoa(port),
		Username:   username,
		Password:   password,
		HTTPClient: createHTTPClient(baseURL, sslVerify),
		SessionTTL: defaultSessionTTL,
	}
}

func NewClientWithToken(baseURL string, port int, token string, sslVerify bool) *Client {
	return &Client{
		BaseURL:       baseURL,
		Port:          strconv.Itoa(port),
		token:         token,
		isStaticToken: true,
		HTTPClient:    createHTTPClient(baseURL, sslVerify),
	}
}

func createHTTPClient(baseURL string, sslVerify bool) *http.Client {
	if warnInsecureTLS(baseURL, sslVerify) {
		log.Warn().Msg("TECHNITIUM_SSL_VERIFY is false: the Technitium server certificate will not be verified")
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			// Verification is on unless the operator turns it off with
			// TECHNITIUM_SSL_VERIFY=false, which exists for in-cluster
			// Technitium servers that use self-signed certificates. The
			// warning above makes that choice visible at startup.
			InsecureSkipVerify: !sslVerify, // #nosec G402 -- opt-out controlled by TECHNITIUM_SSL_VERIFY
		},
	}
	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}
}

func (c *Client) Login() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.loginLocked()
}

func (c *Client) loginLocked() error {
	if c.isStaticToken {
		return fmt.Errorf("login disabled: client is configured with a static API token")
	}

	metrics.TotalApiCalls.Inc()
	timer := prometheus.NewTimer(metrics.ApiCallLatency.WithLabelValues("login"))
	defer timer.ObserveDuration()

	reqURL := fmt.Sprintf("%s:%s%s", c.BaseURL, c.Port, "/api/user/login")
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		metrics.FailedApiCallsTotal.Inc()
		return fmt.Errorf("failed to create login request: %w", err)
	}

	q := req.URL.Query()
	q.Add("user", c.Username)
	q.Add("pass", c.Password)
	req.URL.RawQuery = q.Encode()

	log.Debug().Str("path", "/api/user/login").Msg("logging in to technitium")
	start := time.Now()
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		metrics.FailedApiCallsTotal.Inc()
		err = redactErr(err)
		log.Error().Err(err).Dur("duration", time.Since(start)).Msg("technitium login request failed")
		return fmt.Errorf("login request failed: %w", err)
	}
	log.Debug().Int("status", resp.StatusCode).Dur("duration", time.Since(start)).Msg("technitium login responded")
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			metrics.FailedApiCallsTotal.Inc()
			log.Error().Msgf("Failed to close response body: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		metrics.FailedApiCallsTotal.Inc()
		return fmt.Errorf("failed to read login response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		metrics.FailedApiCallsTotal.Inc()
		return fmt.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		metrics.FailedApiCallsTotal.Inc()
		return fmt.Errorf("failed to parse login JSON: %w", err)
	}

	if apiResp.Status != "ok" {
		metrics.FailedApiCallsTotal.Inc()
		return fmt.Errorf("login failed: %s", apiResp.ErrorMessage)
	}

	if apiResp.Token == "" {
		metrics.FailedApiCallsTotal.Inc()
		return fmt.Errorf("login succeeded but no token was returned by the server")
	}

	c.token = apiResp.Token
	c.renewSessionLocked()
	log.Info().Dur("sessionTTL", c.SessionTTL).Time("expires", c.tokenExpiry).Msg("logged in to technitium; reusing this session until it expires or is rejected")

	return nil
}

func (c *Client) DoRequest(method, path string, params url.Values) ([]byte, error) {
	startTime := time.Now()
	timer := prometheus.NewTimer(metrics.ApiCallLatency.WithLabelValues(path))
	duration := time.Since(startTime)
	defer timer.ObserveDuration()

	c.mu.Lock()
	if !c.isStaticToken && c.sessionNeedsLoginLocked() {
		if err := c.loginLocked(); err != nil {
			c.mu.Unlock()
			return nil, fmt.Errorf("auto-login failed: %w", err)
		}
	}
	currentToken := c.token
	c.mu.Unlock()

	if params == nil {
		params = url.Values{}
	}
	params.Set("token", currentToken)

	reqURL := fmt.Sprintf("%s:%s%s", c.BaseURL, c.Port, path)
	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		metrics.FailedApiCallsTotal.Inc()
		return nil, fmt.Errorf("failed to create API request: %w", err)
	}
	req.URL.RawQuery = params.Encode()

	log.Trace().Str("method", method).Str("path", path).Str("url", redactURL(req.URL.String())).Msg("calling technitium API")
	callStart := time.Now()
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		metrics.FailedApiCallsTotal.Inc()
		err = redactErr(err)
		log.Error().Err(err).Str("path", path).Dur("duration", time.Since(callStart)).Msg("technitium API request failed")
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	log.Debug().Str("path", path).Int("status", resp.StatusCode).Dur("duration", time.Since(callStart)).Msg("technitium API responded")
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			metrics.FailedApiCallsTotal.Inc()
			log.Error().Msgf("Failed to close response body: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		metrics.FailedApiCallsTotal.Inc()
		return nil, fmt.Errorf("failed to read API response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		metrics.TotalApiCalls.Inc()
		metrics.ApiCallLatency.WithLabelValues(path).Observe(duration.Seconds())
		if err := c.checkAPIStatus(path, body); err != nil {
			return nil, err
		}
		if !c.isStaticToken {
			// Technitium's session timeout counts from the last use, so a
			// successful call pushes the client-side expiry out as well.
			c.mu.Lock()
			c.renewSessionLocked()
			c.mu.Unlock()
		}

	case http.StatusUnauthorized, http.StatusForbidden:
		metrics.FailedApiCallsTotal.Inc()
		metrics.ApiCallLatency.WithLabelValues(path).Observe(duration.Seconds())
		if !c.isStaticToken {
			c.mu.Lock()
			c.token = ""
			c.tokenExpiry = time.Time{}
			c.mu.Unlock()
		}
		return nil, fmt.Errorf("authentication rejected (status %d): %s", resp.StatusCode, string(body))

	default:
		metrics.FailedApiCallsTotal.Inc()
		metrics.ApiCallLatency.WithLabelValues(path).Observe(duration.Seconds())
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// checkAPIStatus reads the Technitium response envelope. Technitium answers
// HTTP 200 for failed calls and reports the outcome in "status", so anything
// other than "ok" is returned as an *APIError (issue #29). On "invalid-token"
// a session client drops its token so the next call logs in again; a static
// token cannot be refreshed, so the rejection is only reported.
func (c *Client) checkAPIStatus(path string, body []byte) error {
	var envelope APIResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		metrics.FailedApiCallsTotal.Inc()
		log.Error().Err(err).Str("path", path).Int("bytes", len(body)).Msg("technitium API returned a body that is not JSON")
		return fmt.Errorf("technitium %s returned a body that is not JSON: %w", path, err)
	}

	if envelope.Status == apiStatusOK {
		log.Trace().Str("path", path).Msg("technitium API status ok")
		return nil
	}

	metrics.FailedApiCallsTotal.Inc()
	apiErr := &APIError{Path: path, Status: envelope.Status, Message: envelope.ErrorMessage}

	if envelope.Status == apiStatusInvalidToken {
		if c.isStaticToken {
			log.Error().Str("path", path).Str("status", envelope.Status).Msg("technitium rejected the static API token; check TECHNITIUM_TOKEN")
			return apiErr
		}
		c.mu.Lock()
		c.token = ""
		c.tokenExpiry = time.Time{}
		c.mu.Unlock()
		log.Warn().Str("path", path).Str("status", envelope.Status).Msg("technitium session token rejected; it was dropped and the next call logs in again")
		return apiErr
	}

	log.Error().Str("path", path).Str("status", envelope.Status).Str("error", envelope.ErrorMessage).Msg("technitium API call failed")
	return apiErr
}

// sessionNeedsLoginLocked reports whether a session client must log in before
// the next call: it has no token, or the token's lifetime has passed. A zero
// expiry with a token set means the lifetime is unlimited (TTL 0). The caller
// holds c.mu.
func (c *Client) sessionNeedsLoginLocked() bool {
	if c.token == "" {
		log.Debug().Msg("no technitium session token; logging in")
		return true
	}
	if !c.tokenExpiry.IsZero() && time.Now().After(c.tokenExpiry) {
		log.Debug().Time("expired", c.tokenExpiry).Msg("technitium session lifetime passed; logging in again")
		return true
	}
	return false
}

// renewSessionLocked moves the session expiry to one lifetime from now, or
// clears it when the lifetime is unlimited. The caller holds c.mu.
func (c *Client) renewSessionLocked() {
	if lifetime := sessionLifetime(c.SessionTTL); lifetime > 0 {
		c.tokenExpiry = time.Now().Add(lifetime)
		return
	}
	c.tokenExpiry = time.Time{}
}
