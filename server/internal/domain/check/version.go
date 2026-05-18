package check

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

func (c Check) Hash() string {
	payload := struct {
		Type             Type                     `json:"type"`
		Target           string                   `json:"target"`
		IntervalSeconds  int32                    `json:"intervalSeconds"`
		PingConfig       *domainping.Config       `json:"pingConfig"`
		TracerouteConfig *domaintraceroute.Config `json:"tracerouteConfig"`
	}{
		Type:             c.Type,
		Target:           c.Target,
		IntervalSeconds:  c.IntervalSeconds,
		PingConfig:       nil,
		TracerouteConfig: nil,
	}

	if c.PingConfig != nil {
		payload.PingConfig = c.PingConfig
	}
	if c.TracerouteConfig != nil {
		payload.TracerouteConfig = c.TracerouteConfig
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
