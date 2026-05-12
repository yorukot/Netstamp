package project

import (
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type projectOutput struct {
	Body projectOutputBody
}

type projectOutputBody struct {
	Project domainproject.Project `json:"project"`
}

type memberOutput struct {
	Body memberOutputBody
}

type memberOutputBody struct {
	Member domainproject.Member `json:"member"`
}
