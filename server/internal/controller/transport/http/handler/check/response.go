package check

import domaincheck "github.com/yorukot/netstamp/internal/domain/check"

type checkOutput struct {
	Body checkOutputBody
}

type checkOutputBody struct {
	Check domaincheck.Check `json:"check"`
}
