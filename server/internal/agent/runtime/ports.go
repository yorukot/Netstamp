package runtime

import "context"

type RuntimeClient interface {
	Hello(ctx context.Context) (HelloResponse, error)
	Heartbeat(ctx context.Context, input HeartbeatInput) (HeartbeatResponse, error)
	ListAssignments(ctx context.Context) (AssignmentsResponse, error)
	SubmitResults(ctx context.Context, input SubmitResultsInput) (SubmitResultsResponse, error)
}

type HeartbeatStatusProvider interface {
	Status() HeartbeatInput
}
