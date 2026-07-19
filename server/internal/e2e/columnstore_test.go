//go:build integration

package e2e

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestTimescaleColumnstorePolicies(t *testing.T) {
	suite := newAPISuite(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	targets := []struct {
		name          string
		orderbyTokens []string
	}{
		{name: "ping_results", orderbyTokens: []string{"started_atdesc"}},
		{name: "tcp_results", orderbyTokens: []string{"started_atdesc"}},
		{name: "http_results", orderbyTokens: []string{"started_atdesc"}},
		{name: "traceroute_results", orderbyTokens: []string{"started_atdesc"}},
		{name: "traceroute_result_hops", orderbyTokens: []string{"started_atdesc", "hop_indexasc"}},
		{name: "traceroute_sampled_runs_1m", orderbyTokens: []string{"bucketdesc"}},
		{name: "ping_result_rollups_1m", orderbyTokens: []string{"bucketdesc"}},
		{name: "tcp_result_rollups_1m", orderbyTokens: []string{"bucketdesc"}},
		{name: "http_result_rollups_1m", orderbyTokens: []string{"bucketdesc"}},
	}

	for _, target := range targets {
		var segmentby string
		var orderby string
		if err := suite.pool.QueryRow(ctx, `
			SELECT segmentby, orderby
			FROM timescaledb_information.hypertable_columnstore_settings
			WHERE hypertable = coalesce(
				(
					SELECT format('%I.%I', materialization_hypertable_schema, materialization_hypertable_name)::regclass
					FROM timescaledb_information.continuous_aggregates
					WHERE view_schema = 'public' AND view_name = $1
				),
				$1::regclass
			)
		`, target.name).Scan(&segmentby, &orderby); err != nil {
			t.Fatalf("query columnstore settings for %s: %v", target.name, err)
		}

		if normalized := normalizeColumnstoreSetting(segmentby); normalized != "probe_id,check_id" {
			t.Errorf("expected %s to segment by probe_id and check_id, got %q", target.name, segmentby)
		}
		normalizedOrder := normalizeColumnstoreSetting(orderby)
		for _, token := range target.orderbyTokens {
			if !strings.Contains(normalizedOrder, token) {
				t.Errorf("expected %s order setting %q to contain %q", target.name, orderby, token)
			}
		}
	}

	var totalPolicies int
	var rawPolicies int
	var longTermPolicies int
	var fixedSchedules int
	var distinctStarts int
	if err := suite.pool.QueryRow(ctx, `
		SELECT
			count(*)::integer,
			count(*) FILTER (
				WHERE schedule_interval = INTERVAL '6 hours'
				  AND config->>'compress_after' = '1 day'
			)::integer,
			count(*) FILTER (
				WHERE schedule_interval = INTERVAL '12 hours'
				  AND config->>'compress_after' = '7 days'
			)::integer,
			count(*) FILTER (WHERE fixed_schedule)::integer,
			count(DISTINCT initial_start)::integer
		FROM timescaledb_information.jobs
		WHERE proc_name = 'policy_compression'
	`).Scan(&totalPolicies, &rawPolicies, &longTermPolicies, &fixedSchedules, &distinctStarts); err != nil {
		t.Fatalf("query columnstore policies: %v", err)
	}
	if totalPolicies != 9 || rawPolicies != 5 || longTermPolicies != 4 {
		t.Errorf(
			"expected 9 columnstore policies split into 5 raw and 4 long-term policies, got total=%d raw=%d long-term=%d",
			totalPolicies,
			rawPolicies,
			longTermPolicies,
		)
	}
	if fixedSchedules != 9 || distinctStarts != 9 {
		t.Errorf(
			"expected all columnstore policies to have distinct fixed schedules, got fixed=%d distinct_starts=%d",
			fixedSchedules,
			distinctStarts,
		)
	}

	var httpMaterializedOnly bool
	if err := suite.pool.QueryRow(ctx, `
		SELECT materialized_only
		FROM timescaledb_information.continuous_aggregates
		WHERE view_schema = 'public' AND view_name = 'http_result_rollups_1m'
	`).Scan(&httpMaterializedOnly); err != nil {
		t.Fatalf("query HTTP continuous aggregate mode: %v", err)
	}
	if httpMaterializedOnly {
		t.Error("expected HTTP continuous aggregate to preserve real-time aggregation")
	}
}

func normalizeColumnstoreSetting(value string) string {
	return strings.NewReplacer(`"`, "", " ", "").Replace(strings.ToLower(value))
}
