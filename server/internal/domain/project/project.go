package project

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/yorukot/spvalidator"
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleEditor, RoleViewer:
		return true
	default:
		return false
	}
}

type Project struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Slug            string     `json:"slug"`
	CreatedByUserID string     `json:"createdByUserId"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"-"`
}

func VNProjectCreatedByUserID(userID string) (string, error) {
	userID = strings.TrimSpace(userID)

	err := spvalidator.Required(userID)
	if err != nil {
		return "", err
	}

	err = spvalidator.UUID(userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func VNProjectName(name string) (string, error) {
	name = strings.TrimSpace(name)

	err := spvalidator.Min(name, 1)
	if err != nil {
		return "", err
	}

	err = spvalidator.Max(name, 64)
	if err != nil {
		return "", err
	}

	return name, nil
}

func VNProjectSlug(slug string) (string, error) {
	slug = strings.TrimSpace(slug)

	err := spvalidator.Min(slug, 1)
	if err != nil {
		return "", err
	}

	err = spvalidator.Max(slug, 64)
	if err != nil {
		return "", err
	}

	var slugRe = regexp.MustCompile(`^[a-z0-9-]+$`)

	if !slugRe.MatchString(slug) {
		return "", errors.New("invalid slug")
	}

	return slug, nil
}

// VNProjectRef returns the normalized project reference.
// Because uudi is a sub-set of project slug so we can do it in this way.
func VNProjectRef(projectRef string) (string, error) {
	return VNProjectSlug(projectRef)
}

type Member struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	UserID    string    `json:"userId"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func VNProjectMemberUserID(userID string) (string, error) {
	err := spvalidator.Required(userID)
	if err != nil {
		return "", err
	}

	err = spvalidator.UUID(userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func VNProjectMemberRole(role Role) (Role, error) {
	role = Role(strings.TrimSpace(string(role)))
	if !role.IsValid() {
		return "", errors.New("invalid role")
	}

	return role, nil
}
