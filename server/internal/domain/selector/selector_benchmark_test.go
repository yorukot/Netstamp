package selector

import (
	"encoding/json"
	"fmt"
	"testing"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
)

var benchmarkMatchCount int
var benchmarkCanonicalSelector json.RawMessage

type benchmarkProbe struct {
	labels []domainlabel.Label
}

func BenchmarkMatchesAcrossProbes(b *testing.B) {
	tests := []struct {
		name       string
		raw        json.RawMessage
		probeCount int
		labelCount int
	}{
		{
			name:       "match_all_1000_probes_5_labels",
			raw:        json.RawMessage(`{}`),
			probeCount: 1000,
			labelCount: 5,
		},
		{
			name:       "eq_half_match_1000_probes_5_labels",
			raw:        json.RawMessage(`{"label":{"key":"region","op":"eq","value":"tokyo"}}`),
			probeCount: 1000,
			labelCount: 5,
		},
		{
			name: "nested_half_match_1000_probes_10_labels",
			raw: json.RawMessage(`{"all":[
				{"any":[
					{"label":{"key":"region","op":"eq","value":"tokyo"}},
					{"label":{"key":"region","op":"eq","value":"osaka"}}
				]},
				{"label":{"key":"env","op":"eq","value":"prod"}},
				{"not":{"label":{"key":"disabled","op":"exists"}}}
			]}`),
			probeCount: 1000,
			labelCount: 10,
		},
		{
			name: "any_missing_worst_case_1000_probes_20_labels",
			raw: json.RawMessage(`{"any":[
				{"label":{"key":"missing_a","op":"exists"}},
				{"label":{"key":"missing_b","op":"exists"}},
				{"label":{"key":"missing_c","op":"exists"}},
				{"label":{"key":"missing_d","op":"exists"}}
			]}`),
			probeCount: 1000,
			labelCount: 20,
		},
		{
			name:       "eq_half_match_10000_probes_5_labels",
			raw:        json.RawMessage(`{"label":{"key":"region","op":"eq","value":"tokyo"}}`),
			probeCount: 10000,
			labelCount: 5,
		},
		{
			name: "any_missing_worst_case_10000_probes_20_labels",
			raw: json.RawMessage(`{"any":[
				{"label":{"key":"missing_a","op":"exists"}},
				{"label":{"key":"missing_b","op":"exists"}},
				{"label":{"key":"missing_c","op":"exists"}},
				{"label":{"key":"missing_d","op":"exists"}}
			]}`),
			probeCount: 10000,
			labelCount: 20,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			selector := mustParseBenchmarkSelector(b, tt.raw)
			probes := benchmarkProbes(tt.probeCount, tt.labelCount)

			b.ReportAllocs()
			b.ResetTimer()

			totalMatches := 0
			for i := 0; i < b.N; i++ {
				matches := 0
				for _, probe := range probes {
					if selector.Matches(probe.labels) {
						matches++
					}
				}
				totalMatches += matches
			}

			b.ReportMetric(float64(tt.probeCount), "probes/op")
			b.ReportMetric(float64(tt.labelCount), "labels/probe")
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*tt.probeCount), "ns/probe")

			benchmarkMatchCount = totalMatches
		})
	}
}

func BenchmarkParseAndCanonicalJSON(b *testing.B) {
	tests := []struct {
		name string
		raw  json.RawMessage
	}{
		{
			name: "match_all",
			raw:  json.RawMessage(`{}`),
		},
		{
			name: "single_label_eq",
			raw:  json.RawMessage(`{"label":{"key":"region","op":"eq","value":"tokyo"}}`),
		},
		{
			name: "nested",
			raw: json.RawMessage(`{"all":[
				{"any":[
					{"label":{"key":"region","op":"eq","value":"tokyo"}},
					{"label":{"key":"region","op":"eq","value":"osaka"}}
				]},
				{"label":{"key":"env","op":"eq","value":"prod"}},
				{"not":{"label":{"key":"disabled","op":"exists"}}}
			]}`),
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()

			var canonical json.RawMessage
			for i := 0; i < b.N; i++ {
				selector, err := Parse(tt.raw)
				if err != nil {
					b.Fatalf("parse selector: %v", err)
				}
				canonical = selector.CanonicalJSON()
			}

			benchmarkCanonicalSelector = canonical
		})
	}
}

func benchmarkProbes(count, labelCount int) []benchmarkProbe {
	probes := make([]benchmarkProbe, count)
	for i := range count {
		probes[i] = benchmarkProbe{labels: benchmarkLabels(i, labelCount)}
	}

	return probes
}

func benchmarkLabels(probeIndex, count int) []domainlabel.Label {
	if count < 2 {
		count = 2
	}

	probeLabels := make([]domainlabel.Label, 0, count)
	if probeIndex%2 == 0 {
		probeLabels = append(probeLabels, domainlabel.Label{Key: "region", Value: "tokyo"})
	} else {
		probeLabels = append(probeLabels, domainlabel.Label{Key: "region", Value: "seoul"})
	}
	if probeIndex%3 == 0 {
		probeLabels = append(probeLabels, domainlabel.Label{Key: "env", Value: "prod"})
	} else {
		probeLabels = append(probeLabels, domainlabel.Label{Key: "env", Value: "staging"})
	}

	for i := len(probeLabels); i < count; i++ {
		probeLabels = append(probeLabels, domainlabel.Label{
			Key:   fmt.Sprintf("key_%02d", i),
			Value: fmt.Sprintf("value_%02d_%02d", probeIndex%16, i),
		})
	}

	return probeLabels
}

func mustParseBenchmarkSelector(b *testing.B, raw json.RawMessage) Selector {
	b.Helper()

	selector, err := Parse(raw)
	if err != nil {
		b.Fatalf("parse selector: %v", err)
	}

	return selector
}
