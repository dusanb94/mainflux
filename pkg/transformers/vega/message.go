package vega

// Message represents Device message.
type Message struct {
	D Content `json:"d,omitempty"`
}

// Content represents Message content.
type Content struct {
	Dev      string  `json:"dev,omitempty"`
	IMEI     string  `json:"IMEI,omitempty"`
	ICCID    string  `json:"ICCID,omitempty"`
	SN       string  `json:"SN,omitempty"`
	Isnp     string  `json:"Isnp,omitempty"`
	Num      int     `json:"num,omitempty"`
	MUTC     float64 `json:"mUTC,omitempty"`
	Reason   string  `json:"reason,omitempty"`
	DUTC     int     `json:"dUTC,omitempty"`
	Bat      int     `json:"bat,omitempty"`
	Temp     float64 `json:"temp,omitempty"`
	Water    string  `json:"water,omitempty"`
	SMagnet  int     `json:"s_magnet,omitempty"`
	SBlocked int     `json:"s_blocked,omitempty"`
	SLeakage int     `json:"s_leakage,omitempty"`
	SBlowout int     `json:"s_blowout,omitempty"`
}
