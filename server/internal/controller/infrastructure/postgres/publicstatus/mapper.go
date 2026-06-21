package pgpublicstatus

import (
	"time"

	"github.com/google/uuid"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

func mapPage(row sqlc.PublicStatusPage) domainpublic.Page {
	return domainpublic.Page{
		ID:                row.ID.String(),
		ProjectID:         row.ProjectID.String(),
		Slug:              row.Slug,
		Title:             row.Title,
		Description:       row.Description,
		Enabled:           row.Enabled,
		DefaultChartMode:  domainpublic.ChartMode(row.DefaultChartMode),
		DefaultChartRange: domainpublic.ChartRange(row.DefaultChartRange),
		CreatedByUserID:   row.CreatedByUserID.String(),
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
		DeletedAt:         row.DeletedAt,
	}
}

func mapElement(row sqlc.GetPublicStatusPageElementRow) domainpublic.Element {
	return newElement(row.ID, row.PublicPageID, row.ProjectID, row.ParentElementID, row.Kind, row.CheckID, row.AssignmentSelectionMode, row.Title, row.Description, row.SortOrder, row.ChartMode, row.ChartRange, row.CreatedAt, row.UpdatedAt)
}

func mapCreatedElement(row sqlc.CreatePublicStatusPageElementRow) domainpublic.Element {
	return newElement(row.ID, row.PublicPageID, row.ProjectID, row.ParentElementID, row.Kind, row.CheckID, row.AssignmentSelectionMode, row.Title, row.Description, row.SortOrder, row.ChartMode, row.ChartRange, row.CreatedAt, row.UpdatedAt)
}

func mapUpdatedElement(row sqlc.UpdatePublicStatusPageElementRow) domainpublic.Element {
	return newElement(row.ID, row.PublicPageID, row.ProjectID, row.ParentElementID, row.Kind, row.CheckID, row.AssignmentSelectionMode, row.Title, row.Description, row.SortOrder, row.ChartMode, row.ChartRange, row.CreatedAt, row.UpdatedAt)
}

func newElement(
	id uuid.UUID,
	publicPageID uuid.UUID,
	projectID uuid.UUID,
	parentElementID *uuid.UUID,
	kind sqlc.PublicStatusElementKind,
	checkID *uuid.UUID,
	assignmentSelectionMode *sqlc.PublicStatusAssignmentSelectionMode,
	title *string,
	description *string,
	sortOrder int32,
	chartMode sqlc.PublicStatusChartMode,
	chartRangeValue *sqlc.PublicStatusChartRange,
	createdAt time.Time,
	updatedAt time.Time,
) domainpublic.Element {
	return domainpublic.Element{
		ID:                      id.String(),
		PublicPageID:            publicPageID.String(),
		ProjectID:               projectID.String(),
		ParentElementID:         stringUUID(parentElementID),
		Kind:                    domainpublic.ElementKind(kind),
		CheckID:                 stringUUID(checkID),
		AssignmentSelectionMode: assignmentSelectionModeValue(assignmentSelectionMode),
		Title:                   title,
		Description:             description,
		SortOrder:               sortOrder,
		ChartMode:               domainpublic.ChartMode(chartMode),
		ChartRange:              chartRange(chartRangeValue),
		CreatedAt:               createdAt,
		UpdatedAt:               updatedAt,
	}
}

func mapListElement(row sqlc.ListPublicStatusPageElementsRow) domainpublic.Element {
	element := domainpublic.Element{
		ID:                      row.ID.String(),
		PublicPageID:            row.PublicPageID.String(),
		ProjectID:               row.ProjectID.String(),
		ParentElementID:         stringUUID(row.ParentElementID),
		Kind:                    domainpublic.ElementKind(row.Kind),
		CheckID:                 stringUUID(row.CheckID),
		AssignmentSelectionMode: assignmentSelectionModeValue(row.AssignmentSelectionMode),
		Title:                   row.Title,
		Description:             row.Description,
		SortOrder:               row.SortOrder,
		ChartMode:               domainpublic.ChartMode(row.ChartMode),
		ChartRange:              chartRange(row.ChartRange),
		CreatedAt:               row.CreatedAt,
		UpdatedAt:               row.UpdatedAt,
		CheckName:               row.CheckName,
		CheckTarget:             row.CheckTarget,
		CheckDescription:        row.CheckDescription,
		CheckIntervalSeconds:    row.CheckIntervalSeconds,
	}
	if row.CheckType != nil {
		checkType := domaincheck.Type(*row.CheckType)
		element.CheckType = &checkType
	}
	return element
}

func mapAssignment(row sqlc.ListPublicStatusAssignmentsRow) domainpublic.Assignment {
	return domainpublic.Assignment{
		ElementID:         row.ElementID.String(),
		AssignmentID:      row.AssignmentID.String(),
		CheckID:           row.CheckID.String(),
		CheckName:         row.CheckName,
		CheckType:         domaincheck.Type(row.CheckType),
		CheckTarget:       row.CheckTarget,
		IntervalSeconds:   row.IntervalSeconds,
		ProbeID:           row.ProbeID.String(),
		ProbeName:         row.ProbeName,
		ProbeLocationName: row.ProbeLocationName,
		LatestStartedAt:   row.LatestStartedAt,
		LatestStatus:      row.LatestStatus,
		LatencyAvgMs:      row.LatencyAvgMs,
		LossPercent:       row.LossPercent,
		ConnectAvgMs:      row.ConnectAvgMs,
		FailurePercent:    row.FailurePercent,
	}
}

func mapIncident(row sqlc.ListPublicStatusIncidentsRow) domainpublic.Incident {
	probeID := row.ProbeID.String()
	return domainpublic.Incident{
		ID:              row.ID.String(),
		CheckID:         row.CheckID.String(),
		CheckName:       row.CheckName,
		ProbeID:         &probeID,
		Status:          string(row.Status),
		Severity:        string(row.Severity),
		OpenedAt:        row.OpenedAt,
		ResolvedAt:      row.ResolvedAt,
		LastTriggeredAt: row.LastTriggeredAt,
		LastValue:       row.LastValue,
		LastSummary:     append([]byte{}, row.LastSummary...),
	}
}

func sqlcChartMode(value domainpublic.ChartMode) sqlc.PublicStatusChartMode {
	return sqlc.PublicStatusChartMode(value)
}

func sqlcChartRange(value domainpublic.ChartRange) sqlc.PublicStatusChartRange {
	return sqlc.PublicStatusChartRange(value)
}

func sqlcElementKind(value domainpublic.ElementKind) sqlc.PublicStatusElementKind {
	return sqlc.PublicStatusElementKind(value)
}

func sqlcAssignmentSelectionMode(value *domainpublic.AssignmentSelectionMode) *sqlc.PublicStatusAssignmentSelectionMode {
	if value == nil {
		return nil
	}
	output := sqlc.PublicStatusAssignmentSelectionMode(*value)
	return &output
}

func chartRange(value *sqlc.PublicStatusChartRange) *domainpublic.ChartRange {
	if value == nil {
		return nil
	}
	output := domainpublic.ChartRange(*value)
	return &output
}

func sqlcOptionalChartRange(value *domainpublic.ChartRange) *sqlc.PublicStatusChartRange {
	if value == nil {
		return nil
	}
	output := sqlc.PublicStatusChartRange(*value)
	return &output
}

func assignmentSelectionModeValue(value *sqlc.PublicStatusAssignmentSelectionMode) *domainpublic.AssignmentSelectionMode {
	if value == nil {
		return nil
	}
	output := domainpublic.AssignmentSelectionMode(*value)
	return &output
}

func stringUUID(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	output := value.String()
	return &output
}
