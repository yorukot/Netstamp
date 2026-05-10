package controlplane

import (
	"context"

	"github.com/yorukot/netstamp/internal/contracts/probecontrol"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) PollAssignments(ctx context.Context) (probecontrol.AssignmentSet, error) {
	if err := ctx.Err(); err != nil {
		return probecontrol.AssignmentSet{}, err
	}
	return probecontrol.AssignmentSet{}, nil
}

func (c *Client) SubmitResults(ctx context.Context, batch probecontrol.ResultBatch) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_ = batch
	return nil
}
