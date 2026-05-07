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
		HTTPClient: createHTTPClient(sslVerify),
	}
}

func NewClientWithToken(baseURL string, port int, token string, sslVerify bool) *Client {
	return &Client{
		BaseURL:       baseURL,
		Port:          strconv.Itoa(port),
		token:         token,
		isStaticToken: true,
		HTTPClient:    createHTTPClient(sslVerify),
	}
}

func createHTTPClient(sslVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !sslVerify,
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

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		metrics.FailedApiCallsTotal.Inc()
		return fmt.Errorf("login request failed: %w", err)
	}
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
	c.tokenExpiry = time.Now().Add(sessionBuffer())

	return nil
}

func (c *Client) DoRequest(method, path string, params url.Values) ([]byte, error) {
	startTime := time.Now()
	timer := prometheus.NewTimer(metrics.ApiCallLatency.WithLabelValues(path))
	duration := time.Since(startTime)
	defer timer.ObserveDuration()

	c.mu.Lock()
	if !c.isStaticToken && (c.token == "" || time.Now().After(c.tokenExpiry)) {
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

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		metrics.FailedApiCallsTotal.Inc()
		return nil, fmt.Errorf("API request failed: %w", err)
	}
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
		if !c.isStaticToken {
			c.mu.Lock()
			c.tokenExpiry = time.Now().Add(sessionBuffer())
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
