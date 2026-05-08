package check

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type ExecutionSpec struct {
	Type            Type
	Target          string
	IntervalSeconds int32
	PingConfig      PingConfig
}

func CheckVersion(spec ExecutionSpec) string {
	payload := struct {
		Type            Type     `json:"type"`
		Target          string   `json:"target"`
		IntervalSeconds int32    `json:"intervalSeconds"`
		PingConfig      pingSpec `json:"pingConfig"`
	}{
		Type:            spec.Type,
		Target:          spec.Target,
		IntervalSeconds: spec.IntervalSeconds,
		PingConfig: pingSpec{
			PacketCount:     spec.PingConfig.PacketCount,
			PacketSizeBytes: spec.PingConfig.PacketSizeBytes,
			TimeoutMs:       spec.PingConfig.TimeoutMs,
			IPFamily:        spec.PingConfig.IPFamily,
		},
	}

	return hashJSON(payload)
}

func SelectorVersion(selector json.RawMessage) string {
	return hashBytes(selector)
}

type pingSpec struct {
	PacketCount     int32     `json:"packetCount"`
	PacketSizeBytes int32     `json:"packetSizeBytes"`
	TimeoutMs       int32     `json:"timeoutMs"`
	IPFamily        *IPFamily `json:"ipFamily,omitempty"`
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
