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

const (
	RecordTypeA     = "A"
	RecordTypeAAAA  = "AAAA"
	RecordTypeCNAME = "CNAME"
	RecordTypeTXT   = "TXT"
	RecordTypeNS    = "NS"
)

type Zone struct {
	Name string `json:"name"`
}

type ZoneRecord struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	TTL   int    `json:"ttl"`
	RData struct {
		IPAddress  string `json:"ipAddress,omitempty"`  // Used by A, AAAA
		CNAME      string `json:"cname,omitempty"`      // Used by CNAME
		Text       string `json:"text,omitempty"`       // Used by TXT
		NameServer string `json:"nameServer,omitempty"` // Used by NS
	} `json:"rData"`
}

// GetDataValue is a helper to pull the relevant string regardless of record type
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
