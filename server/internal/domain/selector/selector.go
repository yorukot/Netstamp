package selector

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
)

// ErrInvalidInput is returned when selector JSON is not a valid selector AST.
var ErrInvalidInput = errors.New("selector input invalid")

// Selector is a parsed selector expression. Its zero value behaves like the
// empty selector {}, which matches every probe.
type Selector struct {
	root node
}

// node is the internal AST interface. Public callers should use Selector so the
// JSON grammar can evolve without exposing concrete node types.
type node interface {
	Matches(labels []domainlabel.Label) bool
	canonicalValue() any
}

type matchAllSelector struct{}

type logicalSelector struct {
	op       string
	children []node
}

type notSelector struct {
	child node
}

type labelSelector struct {
	key    string
	op     string
	value  string
	values []string
}

type nodeObject map[string]json.RawMessage

// Parse validates selector JSON and builds the internal AST used for matching.
// Empty raw JSON and JSON null are treated as {}.
func Parse(raw json.RawMessage) (Selector, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return Selector{root: matchAllSelector{}}, nil
	}

	object, err := parseNodeObject(raw)
	if err != nil {
		return Selector{}, err
	}

	root, err := parseObject(object)
	if err != nil {
		return Selector{}, err
	}

	return Selector{root: root}, nil
}

// Matches evaluates the selector against a probe's labels.
func (s Selector) Matches(labels []domainlabel.Label) bool {
	return s.rootOrDefault().Matches(labels)
}

// CanonicalJSON returns the backend-owned JSON shape for this selector.
// Callers can persist this form to keep semantically identical selectors stable.
func (s Selector) CanonicalJSON() json.RawMessage {
	return mustMarshalCanonical(s.rootOrDefault().canonicalValue())
}

func (s Selector) rootOrDefault() node {
	if s.root == nil {
		return matchAllSelector{}
	}

	return s.root
}

func parseRawNode(raw json.RawMessage) (node, error) {
	object, err := parseNodeObject(raw)
	if err != nil {
		return nil, err
	}

	return parseObject(object)
}

func parseNodeObject(raw json.RawMessage) (nodeObject, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, ErrInvalidInput
	}

	var object nodeObject
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, ErrInvalidInput
	}

	return object, nil
}

// parseObject dispatches one AST node. Each selector node must have exactly one
// operator so the expression has unambiguous semantics.
func parseObject(object nodeObject) (node, error) {
	if len(object) == 0 {
		return matchAllSelector{}, nil
	}
	if len(object) != 1 {
		return nil, ErrInvalidInput
	}

	var op string
	var raw json.RawMessage
	for op, raw = range object {
		break
	}

	switch op {
	case "all", "any":
		return parseLogical(op, raw)
	case "not":
		return parseNot(raw)
	case "label":
		return parseLabel(raw)
	default:
		return nil, ErrInvalidInput
	}
}

func parseLogical(op string, raw json.RawMessage) (node, error) {
	var childrenRaw []json.RawMessage
	if err := json.Unmarshal(raw, &childrenRaw); err != nil {
		return nil, ErrInvalidInput
	}
	if len(childrenRaw) == 0 {
		return nil, ErrInvalidInput
	}

	children := make([]node, 0, len(childrenRaw))
	for _, childRaw := range childrenRaw {
		child, err := parseRawNode(childRaw)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}

	return logicalSelector{op: op, children: children}, nil
}

func parseNot(raw json.RawMessage) (node, error) {
	child, err := parseRawNode(raw)
	if err != nil {
		return nil, err
	}

	return notSelector{child: child}, nil
}

func parseLabel(raw json.RawMessage) (node, error) {
	object, err := parseNodeObject(raw)
	if err != nil {
		return nil, err
	}

	key, op, err := parseLabelBase(object)
	if err != nil {
		return nil, err
	}

	switch op {
	case "eq":
		return parseStringValueLabel(object, key, op)
	case "in":
		return parseValuesLabel(object, key, op)
	case "exists":
		return parseExistsLabel(object, key, op)
	default:
		return nil, ErrInvalidInput
	}
}

func parseLabelBase(object nodeObject) (string, string, error) {
	keyRaw, ok := object["key"]
	if !ok {
		return "", "", ErrInvalidInput
	}
	opRaw, ok := object["op"]
	if !ok {
		return "", "", ErrInvalidInput
	}

	key, err := parseRequiredString(keyRaw)
	if err != nil {
		return "", "", err
	}
	op, err := parseRequiredString(opRaw)
	if err != nil {
		return "", "", err
	}

	return key, op, nil
}

func parseStringValueLabel(object nodeObject, key, op string) (node, error) {
	if len(object) != 3 {
		return nil, ErrInvalidInput
	}
	raw, ok := object["value"]
	if !ok {
		return nil, ErrInvalidInput
	}
	value, err := parseRequiredString(raw)
	if err != nil {
		return nil, err
	}

	return labelSelector{key: key, op: op, value: value}, nil
}

func parseValuesLabel(object nodeObject, key, op string) (node, error) {
	if len(object) != 3 {
		return nil, ErrInvalidInput
	}
	raw, ok := object["values"]
	if !ok {
		return nil, ErrInvalidInput
	}
	values, err := parseRequiredStringSet(raw)
	if err != nil {
		return nil, err
	}

	return labelSelector{key: key, op: op, values: values}, nil
}

func parseExistsLabel(object nodeObject, key, op string) (node, error) {
	if len(object) != 2 {
		return nil, ErrInvalidInput
	}

	return labelSelector{key: key, op: op}, nil
}

func parseRequiredString(raw json.RawMessage) (string, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", ErrInvalidInput
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrInvalidInput
	}

	return value, nil
}

func parseRequiredStringSet(raw json.RawMessage) ([]string, error) {
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, ErrInvalidInput
	}
	if len(values) == 0 {
		return nil, ErrInvalidInput
	}

	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, ErrInvalidInput
		}
		if slices.Contains(normalized, value) {
			continue
		}
		normalized = append(normalized, value)
	}
	return normalized, nil
}

func (matchAllSelector) Matches([]domainlabel.Label) bool {
	return true
}

func (s logicalSelector) Matches(labels []domainlabel.Label) bool {
	if s.op == "all" {
		for _, child := range s.children {
			if !child.Matches(labels) {
				return false
			}
		}
		return true
	}

	for _, child := range s.children {
		if child.Matches(labels) {
			return true
		}
	}
	return false
}

func (s notSelector) Matches(labels []domainlabel.Label) bool {
	return !s.child.Matches(labels)
}

func (s labelSelector) Matches(labels []domainlabel.Label) bool {
	switch s.op {
	case "eq":
		return s.matchesEq(labels)
	case "in":
		return s.matchesIn(labels)
	case "exists":
		return s.matchesExists(labels)
	default:
		return false
	}
}

func (s labelSelector) matchesEq(labels []domainlabel.Label) bool {
	for _, label := range labels {
		if label.Key == s.key && label.Value == s.value {
			return true
		}
	}

	return false
}

func (s labelSelector) matchesIn(labels []domainlabel.Label) bool {
	for _, label := range labels {
		if label.Key == s.key && slices.Contains(s.values, label.Value) {
			return true
		}
	}

	return false
}

func (s labelSelector) matchesExists(labels []domainlabel.Label) bool {
	for _, label := range labels {
		if label.Key == s.key {
			return true
		}
	}

	return false
}

func (matchAllSelector) canonicalValue() any {
	return struct{}{}
}

func (s logicalSelector) canonicalValue() any {
	children := make([]any, 0, len(s.children))
	for _, child := range s.children {
		children = append(children, child.canonicalValue())
	}

	if s.op == "all" {
		return struct {
			All []any `json:"all"`
		}{All: children}
	}

	return struct {
		Any []any `json:"any"`
	}{Any: children}
}

func (s notSelector) canonicalValue() any {
	return struct {
		Not any `json:"not"`
	}{Not: s.child.canonicalValue()}
}

func (s labelSelector) canonicalValue() any {
	return struct {
		Label canonicalLabel `json:"label"`
	}{Label: canonicalLabel{
		Key:    s.key,
		Op:     s.op,
		Value:  s.canonicalValueField(),
		Values: s.values,
	}}
}

func (s labelSelector) canonicalValueField() any {
	if s.op == "eq" {
		return s.value
	}

	return nil
}

type canonicalLabel struct {
	Key    string   `json:"key"`
	Op     string   `json:"op"`
	Value  any      `json:"value,omitempty"`
	Values []string `json:"values,omitempty"`
}

func mustMarshalCanonical(value any) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	return raw
}
