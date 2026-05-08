package check

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

type ExecutionSpec struct {
	Type            Type
	Target          string
	IntervalSeconds int32
	PingConfig      domainping.Config
}

func CheckVersion(spec ExecutionSpec) string {
	payload := struct {
		Type            Type                      `json:"type"`
		Target          string                    `json:"target"`
		IntervalSeconds int32                     `json:"intervalSeconds"`
		PingConfig      domainping.VersionPayload `json:"pingConfig"`
	}{
		Type:            spec.Type,
		Target:          spec.Target,
		IntervalSeconds: spec.IntervalSeconds,
		PingConfig:      domainping.ConfigVersionPayload(spec.PingConfig),
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
