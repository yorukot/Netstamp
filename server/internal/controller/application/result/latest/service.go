package latest

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	domainresult "github.com/yorukot/netstamp/internal/domain/result"
)

type Service struct {
	latestResults Repository
	projectAccess resultshared.ProjectAccess
}

func NewService(latestResults Repository, projectAccess resultshared.ProjectAccess) *Service {
	return &Service{
		latestResults: latestResults,
		projectAccess: projectAccess,
	}
}

func (s *Service) Query(ctx context.Context, input QueryInput) (Output, error) {
	ctx, span := resultTracer.Start(ctx, "result.latest.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
		attrResultType.String(input.Type),
	))
	defer span.End()

	normalized, err := normalizeQueryInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result latest query input")
		return Output{}, err
	}
	span.SetAttributes(
		attrProjectRef.String(normalized.projectRef),
		attrProbeID.String(normalized.probeID),
		attrCheckID.String(normalized.checkID),
	)
	if normalized.resultType != nil {
		span.SetAttributes(attrResultType.String(*normalized.resultType))
	}

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.projectRef, normalized.currentUserID)
	if err != nil {
		span.SetStatus(codes.Error, "project lookup failed")
		span.RecordError(err)
		return Output{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.latestResults == nil {
		configuredErr := errors.New("latest result repository is not configured")
		span.SetStatus(codes.Error, "latest result repository missing")
		span.RecordError(configuredErr)
		return Output{}, configuredErr
	}

	result, err := s.latestResults.ListLatestResults(ctx, domainresult.LatestResultQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.probeID,
		CheckID:   normalized.checkID,
		Type:      domainLatestResultType(normalized.resultType),
	})
	if err != nil {
		span.SetStatus(codes.Error, "latest result query failed")
		span.RecordError(err)
		return Output{}, err
	}

	return Output{Results: newResults(result.Results)}, nil
}

func newResults(results []domainresult.LatestResult) []Result {
	values := make([]Result, 0, len(results))
	for _, result := range results {
		values = append(values, Result{
			Type:            string(result.Type),
			ProbeID:         result.ProbeID,
			CheckID:         result.CheckID,
			LatestStartedAt: result.LatestStartedAt,
			LatestStatus:    result.LatestStatus,
		})
	}
	return values
}

func domainLatestResultType(value *string) *domainresult.LatestResultType {
	if value == nil {
		return nil
	}
	resultType := domainresult.LatestResultType(*value)
	return &resultType
}
