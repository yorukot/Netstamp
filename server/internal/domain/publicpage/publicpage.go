package publicpage

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/yorukot/spvalidator"
)

var (
	ErrPageNotFound          = errors.New("public page not found")
	ErrFolderNotFound        = errors.New("public page folder not found")
	ErrCheckNotPublished     = errors.New("public page check not published")
	ErrCheckAlreadyPublished = errors.New("public page check already published")
	ErrDuplicateSlug         = errors.New("public page slug already exists")
	ErrInvalidInput          = errors.New("public page input invalid")
)

type Page struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"projectId"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Enabled     bool       `json:"enabled"`
	Folders     []Folder   `json:"folders,omitempty"`
	Pairs       []PingPair `json:"pairs,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"-"`
}

type Folder struct {
	ID          string           `json:"id"`
	PageID      string           `json:"pageId"`
	ParentID    *string          `json:"parentId,omitempty"`
	Name        string           `json:"name"`
	Description *string          `json:"description,omitempty"`
	SortOrder   int32            `json:"sortOrder"`
	Checks      []PublishedCheck `json:"checks,omitempty"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
}

type PublishedCheck struct {
	ID              string     `json:"id"`
	FolderID        string     `json:"folderId"`
	Name            string     `json:"name"`
	Description     *string    `json:"description,omitempty"`
	IntervalSeconds int32      `json:"intervalSeconds"`
	SortOrder       int32      `json:"sortOrder"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"-"`
}

type PingPair struct {
	FolderID             string  `json:"folderId"`
	ProbeID              string  `json:"probeId"`
	ProbeName            string  `json:"probeName"`
	ProbeLocationName    *string `json:"probeLocationName,omitempty"`
	ProbeStatus          string  `json:"probeStatus"`
	CheckID              string  `json:"checkId"`
	CheckName            string  `json:"checkName"`
	CheckDescription     *string `json:"checkDescription,omitempty"`
	CheckIntervalSeconds int32   `json:"checkIntervalSeconds"`
}

type PageUpdate struct {
	ProjectID      string
	ID             string
	Slug           *string
	Title          *string
	Description    *string
	DescriptionSet bool
	Enabled        *bool
}

type FolderUpdate struct {
	ProjectID      string
	PageID         string
	ID             string
	ParentID       *string
	ParentIDSet    bool
	Name           *string
	Description    *string
	DescriptionSet bool
	SortOrder      *int32
}

func VNPageID(pageID string) (string, error) {
	return vnUUID(pageID)
}

func VNFolderID(folderID string) (string, error) {
	return vnUUID(folderID)
}

func VNSlug(slug string) (string, error) {
	slug = strings.TrimSpace(slug)
	if err := spvalidator.Min(slug, 1); err != nil {
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
	if description == nil {
		return nil, nil //nolint:nilnil // Nil means no description was provided.
	}

	trimmed := strings.TrimSpace(*description)
	if trimmed == "" {
		return nil, nil //nolint:nilnil // Nil means an empty description should be stored as absent.
	}
	if err := spvalidator.Max(trimmed, 1024); err != nil {
		return nil, err
	}

	return &trimmed, nil
}

func VNFolderName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if err := spvalidator.Required(name); err != nil {
		return "", err
	}
	if err := spvalidator.Max(name, 128); err != nil {
		return "", err
	}

	return name, nil
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
