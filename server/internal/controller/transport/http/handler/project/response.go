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

type inviteOutput struct {
	Body inviteOutputBody
}

type inviteOutputBody struct {
	Invite domainproject.Invite `json:"invite"`
}

type inviteListOutput struct {
	Body inviteListOutputBody
}

type inviteListOutputBody struct {
	Invites []domainproject.Invite `json:"invites"`
}
