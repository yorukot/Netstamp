package label

import (
	"errors"

	"github.com/yorukot/netstamp/internal/domain/identity"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

var (
	ErrProjectNotFound    = domainproject.ErrProjectNotFound
	ErrUserNotFound       = identity.ErrUserNotFound
	ErrLabelNotFound      = domainlabel.ErrLabelNotFound
	ErrLabelAlreadyExists = domainlabel.ErrLabelAlreadyExists
	ErrInvalidInput       = domainlabel.ErrInvalidInput
	ErrForbidden          = errors.New("label action forbidden")
)
