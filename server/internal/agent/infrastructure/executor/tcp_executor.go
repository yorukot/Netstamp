package executor

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/yorukot/netstamp/internal/agent/scheduling"
	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

type TCPExecutor struct{}

type tcpExecutionError struct {
	status  domaintcp.Status
	code    string
	message string
}

func NewTCPExecutor() *TCPExecutor {
	return &TCPExecutor{}
}

func (e *TCPExecutor) Execute(ctx context.Context, req scheduling.RunRequest) agentworker.ResultEnvelope {
	result := e.execute(ctx, req)
	return agentworker.ResultEnvelope{
		CheckID: req.Check.ID,
		Type:    domaincheck.TypeTCP,
		TCP:     result,
	}
}

func (e *TCPExecutor) execute(ctx context.Context, req scheduling.RunRequest) domaintcp.Result {
	startedAt := req.ScheduledAt.UTC()
	finishedAt := time.Now().UTC()
	if req.Check.TCPConfig == nil {
		return errorTCPResult(startedAt, finishedAt, "missing_tcp_config", "tcp config is missing")
	}

	resolved, connectDurationMs, err := e.run(ctx, req.Check.Target, *req.Check.TCPConfig)
	finishedAt = time.Now().UTC()
	if err != nil {
		var tcpErr *tcpExecutionError
		if errors.As(err, &tcpErr) {
			return tcpResultFromAttempt(startedAt, finishedAt, resolved, connectDurationMs, tcpErr.status, tcpErr.code, tcpErr.message)
		}
		return tcpResultFromAttempt(startedAt, finishedAt, resolved, connectDurationMs, domaintcp.StatusError, "tcp_connect_failed", err.Error())
	}

	return tcpResultFromAttempt(startedAt, finishedAt, resolved, connectDurationMs, domaintcp.StatusSuccessful, "", "")
}

func (e *TCPExecutor) run(ctx context.Context, target string, config domaintcp.Config) (pingTarget, *float64, error) {
	timeout := time.Duration(config.TimeoutMs) * time.Millisecond
	if config.Port <= 0 || timeout <= 0 {
		return pingTarget{}, nil, newTCPExecutionError(domaintcp.StatusError, "invalid_tcp_config", "tcp config contains non-positive values")
	}

	deadline := time.Now().Add(timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}

	resolveCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	resolved, err := resolvePingTarget(resolveCtx, target, config.IPFamily)
	if err != nil {
		if ctxErr := contextTCPExecutionError(resolveCtx); ctxErr != nil {
			return pingTarget{}, nil, ctxErr
		}
		return pingTarget{}, nil, newTCPExecutionError(domaintcp.StatusError, "resolve_failed", err.Error())
	}
	if !time.Now().Before(deadline) {
		return resolved, nil, newTCPExecutionError(domaintcp.StatusTimeout, "tcp_timeout", "connection timed out")
	}

	dialCtx, dialCancel := context.WithDeadline(ctx, deadline)
	defer dialCancel()
	dialer := net.Dialer{Timeout: time.Until(deadline)}
	connectStartedAt := time.Now()
	conn, err := dialer.DialContext(dialCtx, tcpNetwork(resolved.ipFamily), net.JoinHostPort(resolved.addr.String(), strconv.Itoa(int(config.Port))))
	connectDurationMs := durationMs(time.Since(connectStartedAt))
	if conn != nil {
		_ = conn.Close()
	}
	if err != nil {
		if ctxErr := contextTCPExecutionError(dialCtx); ctxErr != nil {
			return resolved, &connectDurationMs, ctxErr
		}
		if isTimeout(err) {
			return resolved, &connectDurationMs, newTCPExecutionError(domaintcp.StatusTimeout, "tcp_timeout", err.Error())
		}
		return resolved, &connectDurationMs, newTCPExecutionError(domaintcp.StatusError, "tcp_connect_failed", err.Error())
	}

	return resolved, &connectDurationMs, nil
}

func tcpNetwork(ipFamily domainnetwork.IPFamily) string {
	if ipFamily == domainnetwork.IPFamilyInet6 {
		return "tcp6"
	}
	return "tcp4"
}

func tcpResultFromAttempt(startedAt, finishedAt time.Time, target pingTarget, connectDurationMs *float64, status domaintcp.Status, errorCode, errorMessage string) domaintcp.Result {
	addr := target.addr
	addrPtr := &addr
	if !addr.IsValid() {
		addrPtr = nil
	}
	ipFamily := target.ipFamily
	ipFamilyPtr := &ipFamily
	if !addr.IsValid() {
		ipFamilyPtr = nil
	}

	return domaintcp.Result{
		StartedAt:         startedAt.UTC(),
		FinishedAt:        finishedAt.UTC(),
		DurationMs:        durationMillis(startedAt, finishedAt),
		Status:            status,
		ConnectDurationMs: connectDurationMs,
		ResolvedIP:        addrPtr,
		IPFamily:          ipFamilyPtr,
		ErrorCode:         optionalString(errorCode),
		ErrorMessage:      optionalString(errorMessage),
	}
}

func errorTCPResult(startedAt, finishedAt time.Time, errorCode, errorMessage string) domaintcp.Result {
	return tcpResultFromAttempt(startedAt, finishedAt, pingTarget{}, nil, domaintcp.StatusError, errorCode, errorMessage)
}

func newTCPExecutionError(status domaintcp.Status, code, message string) *tcpExecutionError {
	return &tcpExecutionError{
		status:  status,
		code:    code,
		message: message,
	}
}

func contextTCPExecutionError(ctx context.Context) *tcpExecutionError {
	err := ctx.Err()
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return newTCPExecutionError(domaintcp.StatusTimeout, "context_deadline_exceeded", err.Error())
	}

	return newTCPExecutionError(domaintcp.StatusError, "context_canceled", err.Error())
}

func (e *tcpExecutionError) Error() string {
	return e.message
}
