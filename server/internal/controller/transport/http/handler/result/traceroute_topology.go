package result

import (
	"context"
	"net/netip"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
)

func (h *Handler) queryTracerouteTopology(ctx context.Context, input *queryTracerouteTopologyInput) (*queryTracerouteTopologyOutput, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.QueryTracerouteTopology(ctx, appresult.QueryTracerouteTopologyInput{
		CurrentUserID: userID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
		CheckID:       input.CheckID,
		FromMs:        optionalInt64(input.From),
		ToMs:          optionalInt64(input.To),
		Limit:         optionalInt32(input.Limit),
	})
	if err != nil {
		return nil, mapResultError(err, "query traceroute topology failed")
	}

	return &queryTracerouteTopologyOutput{Body: newQueryTracerouteTopologyBody(output)}, nil
}

type queryTracerouteTopologyInput struct {
	Ref     string
	ProbeID string
	CheckID string
	From    int64
	To      int64
	Limit   int32
}

type queryTracerouteTopologyOutput struct {
	Body queryTracerouteTopologyBody
}

type queryTracerouteTopologyBody struct {
	Nodes []tracerouteTopologyNodeBody        `json:"nodes"`
	Edges []tracerouteTopologyEdgeBody        `json:"edges"`
	Query tracerouteTopologyQueryMetadataBody `json:"query"`
}

type tracerouteTopologyNodeBody struct {
	ID          string      `json:"id"`
	Kind        string      `json:"kind"`
	Label       string      `json:"label"`
	Address     *netip.Addr `json:"address,omitempty"`
	Hostname    *string     `json:"hostname,omitempty"`
	ProbeID     *string     `json:"probeId,omitempty"`
	CheckID     *string     `json:"checkId,omitempty"`
	Target      *string     `json:"target,omitempty"`
	HopIndex    *int32      `json:"hopIndex,omitempty"`
	SeenCount   int32       `json:"seenCount"`
	AvgRttMs    *float64    `json:"avgRttMs,omitempty"`
	LossPercent *float64    `json:"lossPercent,omitempty"`
}

type tracerouteTopologyEdgeBody struct {
	ID          string   `json:"id"`
	Source      string   `json:"source"`
	Target      string   `json:"target"`
	SeenCount   int32    `json:"seenCount"`
	AvgRttMs    *float64 `json:"avgRttMs,omitempty"`
	LossPercent *float64 `json:"lossPercent,omitempty"`
}

type tracerouteTopologyQueryMetadataBody struct {
	FromMs int64 `json:"from"`
	ToMs   int64 `json:"to"`
	Limit  int32 `json:"limit"`
}

func newQueryTracerouteTopologyBody(output appresult.TracerouteTopologyOutput) queryTracerouteTopologyBody {
	nodes := make([]tracerouteTopologyNodeBody, 0, len(output.Nodes))
	for _, node := range output.Nodes {
		nodes = append(nodes, tracerouteTopologyNodeBody{
			ID:          node.ID,
			Kind:        node.Kind,
			Label:       node.Label,
			Address:     node.Address,
			Hostname:    node.Hostname,
			ProbeID:     node.ProbeID,
			CheckID:     node.CheckID,
			Target:      node.Target,
			HopIndex:    node.HopIndex,
			SeenCount:   node.SeenCount,
			AvgRttMs:    node.AvgRttMs,
			LossPercent: node.LossPercent,
		})
	}

	edges := make([]tracerouteTopologyEdgeBody, 0, len(output.Edges))
	for _, edge := range output.Edges {
		edges = append(edges, tracerouteTopologyEdgeBody{
			ID:          edge.ID,
			Source:      edge.Source,
			Target:      edge.Target,
			SeenCount:   edge.SeenCount,
			AvgRttMs:    edge.AvgRttMs,
			LossPercent: edge.LossPercent,
		})
	}

	return queryTracerouteTopologyBody{
		Nodes: nodes,
		Edges: edges,
		Query: tracerouteTopologyQueryMetadataBody{
			FromMs: output.Query.FromMs,
			ToMs:   output.Query.ToMs,
			Limit:  output.Query.Limit,
		},
	}
}
