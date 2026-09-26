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
	"time"
)

// secretParams are the query parameters that carry credentials in Technitium
// API calls. Their values must never reach a log line or a returned error.
var secretParams = []string{"pass", "token"}

const redacted = "REDACTED"

func sessionBuffer() time.Duration {
	config := StartupConfig{}
	return time.Duration(config.SessionTTL) * time.Minute
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
