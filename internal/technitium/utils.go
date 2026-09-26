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
	"errors"
	"net/url"
	"strings"
	"time"
)

// secretParams are the query parameters that carry credentials in Technitium
// API calls. Their values must never reach a log line or a returned error.
var secretParams = []string{"pass", "token"}

const redacted = "REDACTED"

// defaultSessionTTL matches the TECHNITIUM_SESSION_TTL default and
// Technitium's default user session timeout.
const defaultSessionTTL = 30 * time.Minute

// sessionLifetime is how long a session token is reused before the client
// logs in again: the configured TTL less a 10% margin, so the token is
// replaced before Technitium expires it. A TTL of 0 or less means the client
// never expires the token itself and relies on Technitium rejecting it with
// "invalid-token" (issue #31).
func sessionLifetime(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return 0
	}
	return ttl - ttl/10
}

// warnInsecureTLS reports whether skipping certificate verification actually
// affects this host. It only matters for https; a plain http host has no
// certificate to verify.
func warnInsecureTLS(baseURL string, sslVerify bool) bool {
	return !sslVerify && strings.HasPrefix(strings.ToLower(baseURL), "https://")
}

// redactURL returns raw with the values of credential query parameters
// replaced. A string that does not parse as a URL is returned unchanged.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	changed := false
	for _, k := range secretParams {
		if q.Has(k) {
			q.Set(k, redacted)
			changed = true
		}
	}
	if !changed {
		return raw
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// redactErr scrubs credentials from the request URL that net/http embeds in
// transport errors. external-dns logs provider errors verbatim, so without
// this a failed call would log the password or token (issue #28).
func redactErr(err error) error {
	var ue *url.Error
	if errors.As(err, &ue) {
		ue.URL = redactURL(ue.URL)
	}
	return err
}
