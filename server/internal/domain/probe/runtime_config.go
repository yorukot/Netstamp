package probe

const (
	DefaultRuntimeHeartbeatIntervalSeconds      int32 = 30
	DefaultRuntimeAssignmentPollIntervalSeconds int32 = 30
	DefaultRuntimeMaxConcurrentChecks           int32 = 16
	DefaultRuntimeInitialBackoffSeconds         int32 = 1
	DefaultRuntimeMaxBackoffSeconds             int32 = 30
	DefaultRuntimeMaxAttempts                   int32 = 5
	DefaultMinimumSupportedAgentVersion               = "0.1.0"
	MaxRuntimeResultGroupBatchSize              int   = 100
)

type RuntimeConfig struct {
	HeartbeatIntervalSeconds      int32 `json:"heartbeatIntervalSeconds"`
	AssignmentPollIntervalSeconds int32 `json:"assignmentPollIntervalSeconds"`
	MaxConcurrentChecks           int32 `json:"maxConcurrentChecks"`
	InitialBackoffSeconds         int32 `json:"initialBackoffSeconds"`
	MaxBackoffSeconds             int32 `json:"maxBackoffSeconds"`
	MaxAttempts                   int32 `json:"maxAttempts"`
}

func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		HeartbeatIntervalSeconds:      DefaultRuntimeHeartbeatIntervalSeconds,
		AssignmentPollIntervalSeconds: DefaultRuntimeAssignmentPollIntervalSeconds,
		MaxConcurrentChecks:           DefaultRuntimeMaxConcurrentChecks,
		InitialBackoffSeconds:         DefaultRuntimeInitialBackoffSeconds,
		MaxBackoffSeconds:             DefaultRuntimeMaxBackoffSeconds,
		MaxAttempts:                   DefaultRuntimeMaxAttempts,
	}
}
