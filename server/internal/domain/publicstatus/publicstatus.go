package publicstatus

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/yorukot/spvalidator"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

var (
	ErrInvalidInput     = errors.New("public status page input invalid")
	ErrPageNotFound     = errors.New("public status page not found")
	ErrElementNotFound  = errors.New("public status page element not found")
	ErrSlugAlreadyExist = errors.New("public status page slug already exists")
)

type ChartMode string

const (
	ChartModeInherit ChartMode = "inherit"
	ChartModeOff     ChartMode = "off"
	ChartModeCompact ChartMode = "compact"
)

type ChartRange string

const (
	ChartRange24h ChartRange = "24h"
	ChartRange7d  ChartRange = "7d"
	ChartRange30d ChartRange = "30d"
)

type ElementKind string

const (
	ElementKindFolder          ElementKind = "folder"
	ElementKindAssignmentGroup ElementKind = "assignment_group"
)

type ElementDisplayMode string

const (
	ElementDisplayModeStatus  ElementDisplayMode = "status"
	ElementDisplayModeHistory ElementDisplayMode = "history"
	ElementDisplayModeLatency ElementDisplayMode = "latency"
	ElementDisplayModeMap     ElementDisplayMode = "map"
)

type Theme string

const (
	ThemeLight Theme = "light"
	ThemeDark  Theme = "dark"
	ThemeAuto  Theme = "auto"
)

type AssignmentSelectionMode string

const (
	AssignmentSelectionModeAllCheck            AssignmentSelectionMode = "all_check"
	AssignmentSelectionModeSelectedAssignments AssignmentSelectionMode = "selected_assignments"
)

type Status string

const (
	StatusOperational Status = "operational"
	StatusDegraded    Status = "degraded"
	StatusDown        Status = "down"
	StatusUnknown     Status = "unknown"
)

type Page struct {
	ID                  string
	ProjectID           string
	Slug                string
	Title               string
	Description         *string
	Enabled             bool
	FooterText          *string
	BannerImageURL      *string
	Theme               Theme
	ShowTargets         bool
	ShowProbeNames      bool
	ShowProbeLocations  bool
	ShowIncidentHistory bool
	ShowGeneratedAt     bool
	CustomCSS           *string
	DefaultChartMode    ChartMode
	DefaultChartRange   ChartRange
	CreatedByUserID     string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}

type Element struct {
	ID                      string
	PublicPageID            string
	ProjectID               string
	ParentElementID         *string
	Kind                    ElementKind
	CheckID                 *string
	AssignmentSelectionMode *AssignmentSelectionMode
	AssignmentIDs           []string
	Title                   *string
	Description             *string
	SortOrder               int32
	DisplayMode             ElementDisplayMode
	ChartMode               ChartMode
	ChartRange              *ChartRange
	CreatedAt               time.Time
	UpdatedAt               time.Time
	CheckName               *string
	CheckType               *domaincheck.Type
	CheckTarget             *string
	CheckDescription        *string
	CheckIntervalSeconds    *int32
}

type Assignment struct {
	ElementID         string
	AssignmentID      string
	CheckID           string
	CheckName         string
	CheckType         domaincheck.Type
	CheckTarget       string
	IntervalSeconds   int32
	ProbeID           string
	ProbeName         string
	ProbeLocationName *string
	ProbeLatitude     *float64
	ProbeLongitude    *float64
	LatestStartedAt   time.Time
	LatestStatus      string
	LatencyAvgMs      *float64
	LossPercent       float64
	ConnectAvgMs      *float64
	FailurePercent    *float64
}

type Incident struct {
	ID              string
	CheckID         string
	CheckName       string
	ProbeID         *string
	Status          string
	Severity        string
	OpenedAt        time.Time
	ResolvedAt      *time.Time
	LastTriggeredAt time.Time
	LastValue       *float64
	LastSummary     []byte
}

type RenderedElement struct {
	Element
	Status                Status
	ResolvedChartMode     ChartMode
	ResolvedChartRange    ChartRange
	Children              []RenderedElement
	LatestStartedAt       *time.Time
	LatestStatus          *string
	AssignmentCount       int32
	SuccessfulAssignments int32
	FailingAssignments    int32
	StaleAssignments      int32
	Metrics               *Metrics
	Chart                 *Chart
	Assignments           []Assignment
}

type Metrics struct {
	LatencyAvgMs   *float64
	LossPercent    *float64
	ConnectAvgMs   *float64
	FailurePercent *float64
}

type Chart struct {
	Range  ChartRange
	Series []Series
}

type DailyStatusDay struct {
	Date          time.Time
	Status        Status
	IncidentCount int32
	Severity      *string
}

type Series struct {
	Name   string
	Labels map[string]string
	Unit   string
	Points []SeriesPoint
}

type SeriesPoint struct {
	TimestampMs int64
	Value       float64
}

func VNPageID(pageID string) (string, error) {
	return vnUUID(pageID)
}

func VNElementID(elementID string) (string, error) {
	return vnUUID(elementID)
}

func VNAssignmentID(assignmentID string) (string, error) {
	return vnUUID(assignmentID)
}

func VNSlug(slug string) (string, error) {
	slug = strings.TrimSpace(slug)
	if err := spvalidator.Required(slug); err != nil {
		return "", err
	}
	if err := spvalidator.Max(slug, 64); err != nil {
		return "", err
	}
	if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(slug) {
		return "", errors.New("invalid slug")
	}
	return slug, nil
}

func VNTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if err := spvalidator.Required(title); err != nil {
		return "", err
	}
	if err := spvalidator.Max(title, 128); err != nil {
		return "", err
	}
	return title, nil
}

func VNDescription(description *string) (*string, error) {
	return vnOptionalText(description, 1024)
}

func VNFooterText(value *string) (*string, error) {
	return vnOptionalText(value, 2048)
}

func VNCustomCSS(value *string) (*string, error) {
	return vnOptionalText(value, 65536)
}

func VNBannerImageURL(value *string) (*string, error) {
	normalized, err := vnOptionalText(value, 2048)
	if err != nil || normalized == nil {
		return normalized, err
	}
	parsed, err := url.Parse(*normalized)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, errors.New("must be an absolute http or https URL")
	}
	return normalized, nil
}

func VNTheme(theme Theme) (Theme, error) {
	theme = Theme(strings.TrimSpace(string(theme)))
	switch theme {
	case ThemeLight, ThemeDark, ThemeAuto:
		return theme, nil
	default:
		return "", errors.New("invalid status page theme")
	}
}

func VNElementDisplayMode(mode ElementDisplayMode) (ElementDisplayMode, error) {
	mode = ElementDisplayMode(strings.TrimSpace(string(mode)))
	switch mode {
	case ElementDisplayModeStatus, ElementDisplayModeHistory, ElementDisplayModeLatency, ElementDisplayModeMap:
		return mode, nil
	default:
		return "", errors.New("invalid element display mode")
	}
}

func VNChartMode(mode ChartMode, allowInherit bool) (ChartMode, error) {
	mode = ChartMode(strings.TrimSpace(string(mode)))
	switch mode {
	case ChartModeOff, ChartModeCompact:
		return mode, nil
	case ChartModeInherit:
		if allowInherit {
			return mode, nil
		}
	}
	return "", errors.New("invalid chart mode")
}

func VNChartRange(chartRange ChartRange) (ChartRange, error) {
	chartRange = ChartRange(strings.TrimSpace(string(chartRange)))
	switch chartRange {
	case ChartRange24h, ChartRange7d, ChartRange30d:
		return chartRange, nil
	default:
		return "", errors.New("invalid chart range")
	}
}

func VNElementKind(kind ElementKind) (ElementKind, error) {
	kind = ElementKind(strings.TrimSpace(string(kind)))
	switch kind {
	case ElementKindFolder, ElementKindAssignmentGroup:
		return kind, nil
	default:
		return "", errors.New("invalid element kind")
	}
}

func VNAssignmentSelectionMode(mode AssignmentSelectionMode) (AssignmentSelectionMode, error) {
	mode = AssignmentSelectionMode(strings.TrimSpace(string(mode)))
	switch mode {
	case AssignmentSelectionModeAllCheck, AssignmentSelectionModeSelectedAssignments:
		return mode, nil
	default:
		return "", errors.New("invalid assignment selection mode")
	}
}

func VNSortOrder(sortOrder int32) (int32, error) {
	if err := spvalidator.Min(sortOrder, 0); err != nil {
		return 0, err
	}
	return sortOrder, nil
}

func vnUUID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if err := spvalidator.Required(value); err != nil {
		return "", err
	}
	if err := spvalidator.UUID(value); err != nil {
		return "", err
	}
	return value, nil
}

func vnOptionalText(value *string, maximum int) (*string, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // Nil is the canonical value for omitted optional text.
	}
	trimmed := strings.TrimSpace(*value)
	if err := spvalidator.Required(trimmed); err != nil {
		return nil, err
	}
	if err := spvalidator.Max(trimmed, maximum); err != nil {
		return nil, err
	}
	return &trimmed, nil
}
