package technitium

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
