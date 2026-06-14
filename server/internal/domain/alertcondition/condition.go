package alertcondition

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yorukot/spvalidator"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

var ErrInvalidCondition = errors.New("alert condition invalid")

type Operator string

const (
	TypeMetricThreshold = "metric_threshold"

	OperatorGT  Operator = "gt"
	OperatorGTE Operator = "gte"
	OperatorLT  Operator = "lt"
	OperatorLTE Operator = "lte"
	OperatorEQ  Operator = "eq"

	MetricPingLossPercent  = "ping.loss_percent"
	MetricPingAverageRTTMs = "ping.average_rtt_ms"
	MetricPingMaxRTTMs     = "ping.max_rtt_ms"
	MetricPingSuccessRate  = "ping.success_rate"
	MetricTCPFailurePct    = "tcp.failure_percent"
	MetricTCPAverageConnMs = "tcp.average_connect_ms"
	MetricTCPMaxConnMs     = "tcp.max_connect_ms"
	MetricTCPSuccessRate   = "tcp.success_rate"
)

type Condition struct {
	Type          string   `json:"type"`
	Metric        string   `json:"metric"`
	Operator      Operator `json:"operator"`
	Threshold     float64  `json:"threshold"`
	WindowSeconds int32    `json:"windowSeconds"`
	MinSamples    int32    `json:"minSamples"`
}

type Requirement struct {
	Metric        string
	WindowSeconds int32
	MinSamples    int32
}

type EvaluationState string

const (
	EvaluationStateFiring              EvaluationState = "firing"
	EvaluationStateClear               EvaluationState = "clear"
	EvaluationStateInsufficientSamples EvaluationState = "insufficient_samples"
	EvaluationStateNoData              EvaluationState = "no_data"
)

type MetricSummary struct {
	Metric        string
	WindowStart   time.Time
	WindowEnd     time.Time
	Samples       int64
	Value         float64
	HasValue      bool
	MinSamples    int32
	WindowSeconds int32
}

type Evaluation struct {
	State   EvaluationState
	Metric  string
	Value   float64
	Summary MetricSummary
}

func Parse(raw json.RawMessage) (Condition, error) {
	if len(raw) == 0 {
		return Condition{}, fmt.Errorf("%w: condition is required", ErrInvalidCondition)
	}
	if len(raw) > 16*1024 {
		return Condition{}, fmt.Errorf("%w: condition JSON is too large", ErrInvalidCondition)
	}

	var condition Condition
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&condition); err != nil {
		return Condition{}, fmt.Errorf("%w: %v", ErrInvalidCondition, err)
	}
	if err := decoder.Decode(&struct{}{}); err == nil {
		return Condition{}, fmt.Errorf("%w: condition must contain a single JSON object", ErrInvalidCondition)
	}

	return Validate(condition)
}

func Validate(condition Condition) (Condition, error) {
	condition.Type = strings.TrimSpace(condition.Type)
	condition.Metric = strings.TrimSpace(condition.Metric)
	condition.Operator = Operator(strings.TrimSpace(string(condition.Operator)))

	if condition.Type == "" {
		condition.Type = TypeMetricThreshold
	}
	if condition.Type != TypeMetricThreshold {
		return Condition{}, fmt.Errorf("%w: unsupported condition type", ErrInvalidCondition)
	}
	if !IsSupportedMetric(condition.Metric) {
		return Condition{}, fmt.Errorf("%w: unsupported metric", ErrInvalidCondition)
	}
	if !isSupportedOperator(condition.Operator) {
		return Condition{}, fmt.Errorf("%w: unsupported operator", ErrInvalidCondition)
	}
	if err := spvalidator.Min(condition.WindowSeconds, int32(60)); err != nil {
		return Condition{}, fmt.Errorf("%w: invalid window seconds", ErrInvalidCondition)
	}
	if err := spvalidator.Max(condition.WindowSeconds, int32(86400)); err != nil {
		return Condition{}, fmt.Errorf("%w: invalid window seconds", ErrInvalidCondition)
	}
	if err := spvalidator.Min(condition.MinSamples, int32(1)); err != nil {
		return Condition{}, fmt.Errorf("%w: invalid min samples", ErrInvalidCondition)
	}
	if err := spvalidator.Max(condition.MinSamples, int32(10000)); err != nil {
		return Condition{}, fmt.Errorf("%w: invalid min samples", ErrInvalidCondition)
	}

	return condition, nil
}

func CanonicalJSON(condition Condition) (json.RawMessage, error) {
	condition, err := Validate(condition)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(condition)
	if err != nil {
		return nil, fmt.Errorf("%w: canonicalize condition", ErrInvalidCondition)
	}
	return data, nil
}

func (c Condition) Requirement() Requirement {
	return Requirement{
		Metric:        c.Metric,
		WindowSeconds: c.WindowSeconds,
		MinSamples:    c.MinSamples,
	}
}

func (c Condition) Evaluate(summary MetricSummary) Evaluation {
	if !summary.HasValue || summary.Samples == 0 {
		return Evaluation{State: EvaluationStateNoData, Metric: c.Metric, Summary: summary}
	}
	if summary.Samples < int64(c.MinSamples) {
		return Evaluation{State: EvaluationStateInsufficientSamples, Metric: c.Metric, Value: summary.Value, Summary: summary}
	}
	if compare(summary.Value, c.Operator, c.Threshold) {
		return Evaluation{State: EvaluationStateFiring, Metric: c.Metric, Value: summary.Value, Summary: summary}
	}
	return Evaluation{State: EvaluationStateClear, Metric: c.Metric, Value: summary.Value, Summary: summary}
}

func CompatibleWithCheckType(metric string, checkType domaincheck.Type) bool {
	switch checkType {
	case domaincheck.TypePing:
		return strings.HasPrefix(metric, "ping.")
	case domaincheck.TypeTCP:
		return strings.HasPrefix(metric, "tcp.")
	default:
		return false
	}
}

func IsSupportedMetric(metric string) bool {
	switch metric {
	case MetricPingLossPercent, MetricPingAverageRTTMs, MetricPingMaxRTTMs, MetricPingSuccessRate,
		MetricTCPFailurePct, MetricTCPAverageConnMs, MetricTCPMaxConnMs, MetricTCPSuccessRate:
		return true
	default:
		return false
	}
}

func isSupportedOperator(operator Operator) bool {
	switch operator {
	case OperatorGT, OperatorGTE, OperatorLT, OperatorLTE, OperatorEQ:
		return true
	default:
		return false
	}
}

func compare(value float64, operator Operator, threshold float64) bool {
	switch operator {
	case OperatorGT:
		return value > threshold
	case OperatorGTE:
		return value >= threshold
	case OperatorLT:
		return value < threshold
	case OperatorLTE:
		return value <= threshold
	case OperatorEQ:
		return value == threshold
	default:
		return false
	}
}
