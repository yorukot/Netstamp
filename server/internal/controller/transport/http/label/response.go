package label

import (
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
)

type labelOutput struct {
	Body labelOutputBody
}

type labelOutputBody struct {
	Label domainlabel.Label `json:"label"`
}
