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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	RecordTypeA     = "A"
	RecordTypeAAAA  = "AAAA"
	RecordTypeCNAME = "CNAME"
	RecordTypeTXT   = "TXT"
	RecordTypeNS    = "NS"
)

type APIResponse struct {
	Status       string          `json:"status"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
	Token        string          `json:"token,omitempty"`
	Response     json.RawMessage `json:"response,omitempty"`
}

type Client struct {
	BaseURL    string
	Port       string
	Username   string
	Password   string
	HTTPClient *http.Client

	token         string
	tokenExpiry   time.Time
	isStaticToken bool
	SessionTTL    time.Duration
	mu            sync.Mutex
}

type Zone struct {
	Name string `json:"name"`
}

type ZoneRecord struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	TTL   int    `json:"ttl"`
	RData struct {
		IPAddress  string `json:"ipAddress,omitempty"`
		CNAME      string `json:"cname,omitempty"`
		Text       string `json:"text,omitempty"`
		NameServer string `json:"nameServer,omitempty"`
	} `json:"rData"`
}

// Values of the "status" field in a Technitium API response. Technitium
// answers HTTP 200 for failed calls too, so this field, not the HTTP status,
// says whether a call worked (issue #29).
const (
	apiStatusOK           = "ok"
	apiStatusInvalidToken = "invalid-token"
)

// ErrInvalidToken is wrapped by an APIError whose status is "invalid-token":
// the session expired or the API token was rejected.
var ErrInvalidToken = errors.New("technitium rejected the token")

// APIError is a Technitium API call that returned a status other than "ok".
type APIError struct {
	Path    string
	Status  string
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("technitium %s returned status %q: %s", e.Path, e.Status, e.Message)
}

// Unwrap lets errors.Is(err, ErrInvalidToken) match token rejections.
func (e *APIError) Unwrap() error {
	if e.Status == apiStatusInvalidToken {
		return ErrInvalidToken
	}
	return nil
}

func (r ZoneRecord) GetDataValue() string {
	switch r.Type {
	case RecordTypeA, RecordTypeAAAA:
		return r.RData.IPAddress
	case RecordTypeCNAME:
		return r.RData.CNAME
	case RecordTypeTXT:
		return r.RData.Text
	case RecordTypeNS:
		return r.RData.NameServer
	default:
		return ""
	}
}
