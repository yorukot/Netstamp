package pgassignment

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func (r *AssignmentRepository) EnqueueRefreshJob(ctx context.Context, target domainassignment.RefreshTarget, maxAttempts int32) error {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "assignment_refresh_jobs", "postgres.assignments.enqueue_refresh_job", "INSERT", "INSERT assignment refresh job")
	defer span.End()

	if maxAttempts <= 0 {
		maxAttempts = domainassignment.DefaultRefreshJobMaxAttempts
	}
	projectID, targetType, targetID, err := parseRefreshTarget(target)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}
	normalizedTarget := domainassignment.RefreshTarget{
		ProjectID: projectID.String(),
		Type:      target.Type,
		TargetID:  targetID.String(),
	}

	_, err = r.queries.EnqueueAssignmentRefreshJob(ctx, sqlc.EnqueueAssignmentRefreshJobParams{
		ProjectID:   projectID,
		TargetType:  targetType,
		TargetID:    targetID,
		DedupeKey:   normalizedTarget.DedupeKey(),
		MaxAttempts: maxAttempts,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *AssignmentRepository) ClaimRefreshJobs(ctx context.Context, limit int32, staleBefore time.Time) ([]domainassignment.RefreshJob, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "assignment_refresh_jobs", "postgres.assignments.claim_refresh_jobs", "UPDATE", "CLAIM assignment refresh jobs")
	defer span.End()

	stale := staleBefore.UTC()
	if _, err := r.queries.RecoverStaleAssignmentRefreshJobs(ctx, &stale); err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}
	rows, err := r.queries.ClaimAssignmentRefreshJobs(ctx, limit)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	jobs := make([]domainassignment.RefreshJob, 0, len(rows))
	for _, row := range rows {
		jobs = append(jobs, mapRefreshJob(row))
	}
	return jobs, nil
}

func (r *AssignmentRepository) MarkRefreshJobSucceeded(ctx context.Context, id string, at time.Time) error {
	uuidValue, err := postgres.ParseUUID(id, domainassignment.ErrInvalidInput)
	if err != nil {
		return err
	}
	completedAt := at.UTC()
	return r.queries.MarkAssignmentRefreshJobSucceeded(ctx, sqlc.MarkAssignmentRefreshJobSucceededParams{
		ID:          uuidValue,
		CompletedAt: &completedAt,
	})
}

func (r *AssignmentRepository) MarkRefreshJobRetry(ctx context.Context, id string, nextAttemptAt time.Time, kind, code, message string) error {
	uuidValue, err := postgres.ParseUUID(id, domainassignment.ErrInvalidInput)
	if err != nil {
		return err
	}
	kindPtr, codePtr, messagePtr := &kind, &code, &message
	return r.queries.MarkAssignmentRefreshJobRetry(ctx, sqlc.MarkAssignmentRefreshJobRetryParams{
		ID:            uuidValue,
		NextAttemptAt: nextAttemptAt.UTC(),
		LastErrorKind: kindPtr,
		LastErrorCode: codePtr,
		LastError:     messagePtr,
	})
}

func (r *AssignmentRepository) MarkRefreshJobFailed(ctx context.Context, id, kind, code, message string) error {
	uuidValue, err := postgres.ParseUUID(id, domainassignment.ErrInvalidInput)
	if err != nil {
		return err
	}
	kindPtr, codePtr, messagePtr := &kind, &code, &message
	return r.queries.MarkAssignmentRefreshJobFailed(ctx, sqlc.MarkAssignmentRefreshJobFailedParams{
		ID:            uuidValue,
		LastErrorKind: kindPtr,
		LastErrorCode: codePtr,
		LastError:     messagePtr,
	})
}

func (r *AssignmentRepository) MarkRefreshJobDiscarded(ctx context.Context, id, kind, code, message string) error {
	uuidValue, err := postgres.ParseUUID(id, domainassignment.ErrInvalidInput)
	if err != nil {
		return err
	}
	kindPtr, codePtr, messagePtr := &kind, &code, &message
	return r.queries.MarkAssignmentRefreshJobDiscarded(ctx, sqlc.MarkAssignmentRefreshJobDiscardedParams{
		ID:            uuidValue,
		LastErrorKind: kindPtr,
		LastErrorCode: codePtr,
		LastError:     messagePtr,
	})
}

func parseRefreshTarget(target domainassignment.RefreshTarget) (uuid.UUID, sqlc.AssignmentRefreshTargetType, uuid.UUID, error) {
	projectID, err := postgres.ParseUUID(target.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return uuid.Nil, "", uuid.Nil, err
	}

	switch target.Type {
	case domainassignment.RefreshTargetProject:
		targetID, err := postgres.ParseUUID(target.TargetID, domainproject.ErrProjectNotFound)
		return projectID, sqlc.AssignmentRefreshTargetTypeProject, targetID, err
	case domainassignment.RefreshTargetProbe:
		targetID, err := postgres.ParseUUID(target.TargetID, domainprobe.ErrProbeNotFound)
		return projectID, sqlc.AssignmentRefreshTargetTypeProbe, targetID, err
	case domainassignment.RefreshTargetCheck:
		targetID, err := postgres.ParseUUID(target.TargetID, domaincheck.ErrCheckNotFound)
		return projectID, sqlc.AssignmentRefreshTargetTypeCheck, targetID, err
	case domainassignment.RefreshTargetLabel:
		targetID, err := postgres.ParseUUID(target.TargetID, domainlabel.ErrLabelNotFound)
		return projectID, sqlc.AssignmentRefreshTargetTypeLabel, targetID, err
	default:
		return uuid.Nil, "", uuid.Nil, domainassignment.ErrInvalidInput
	}
}

func mapRefreshJob(row sqlc.AssignmentRefreshJob) domainassignment.RefreshJob {
	return domainassignment.RefreshJob{
		ID:            row.ID.String(),
		ProjectID:     row.ProjectID.String(),
		TargetType:    domainassignment.RefreshTargetType(row.TargetType),
		TargetID:      row.TargetID.String(),
		Status:        domainassignment.RefreshJobStatus(row.Status),
		AttemptCount:  row.AttemptCount,
		MaxAttempts:   row.MaxAttempts,
		NextAttemptAt: row.NextAttemptAt,
		LastAttemptAt: row.LastAttemptAt,
		CompletedAt:   row.CompletedAt,
		LastErrorKind: row.LastErrorKind,
		LastErrorCode: row.LastErrorCode,
		LastError:     row.LastError,
		DedupeKey:     row.DedupeKey,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}
