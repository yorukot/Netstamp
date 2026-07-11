package check

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

func (c Check) Hash() string {
	payload := struct {
		Type             Type                     `json:"type"`
		Target           string                   `json:"target"`
		IntervalSeconds  int32                    `json:"intervalSeconds"`
		PingConfig       *domainping.Config       `json:"pingConfig"`
		TCPConfig        *domaintcp.Config        `json:"tcpConfig"`
		TracerouteConfig *domaintraceroute.Config `json:"tracerouteConfig"`
		HTTPConfig       *domainhttp.Config       `json:"httpConfig"`
	}{
		Type:             c.Type,
		Target:           c.Target,
		IntervalSeconds:  c.IntervalSeconds,
		PingConfig:       nil,
		TCPConfig:        nil,
		TracerouteConfig: nil,
		HTTPConfig:       nil,
	}

	if c.PingConfig != nil {
		payload.PingConfig = c.PingConfig
	}
	if c.TCPConfig != nil {
		payload.TCPConfig = c.TCPConfig
	}
	if c.TracerouteConfig != nil {
		payload.TracerouteConfig = c.TracerouteConfig
	}
	if c.HTTPConfig != nil {
		payload.HTTPConfig = c.HTTPConfig
	}

	return hashJSON(payload)
}

func SelectorVersion(selector json.RawMessage) string {
	return hashBytes(selector)
}

func hashJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	return hashBytes(raw)
}

func hashBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}
