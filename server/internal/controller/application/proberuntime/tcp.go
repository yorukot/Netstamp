package proberuntime

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

type normalizedTCPMetadata struct {
	ipFamily     *domainnetwork.IPFamily
	errorCode    *string
	errorMessage *string
}

func normalizeTCPResult(input TCPResultInput, fieldPrefix string) (domaintcp.ResultStorageInput, error) {
	var validation appvalidation.Collector

	timing, err := normalizeResultTiming(
		input.StartedAt,
		input.FinishedAt,
		input.DurationMs,
		fieldPrefix,
		domaintcp.VNResultDurationMs,
	)
	if err != nil {
		if !validation.AddValidation(err) {
			return domaintcp.ResultStorageInput{}, err
		}
	}
	status, err := normalizeTCPStatus(input.Status, fieldPrefix)
	if err != nil {
		if !validation.AddValidation(err) {
			return domaintcp.ResultStorageInput{}, err
		}
	}
	connectDurationMs, err := domaintcp.VNResultConnectDurationMs(input.ConnectDurationMs)
	if err != nil {
		validation.AddError(resultField(fieldPrefix, "connectDurationMs"), err, input.ConnectDurationMs)
	}
	metadata, err := normalizeTCPMetadata(input, fieldPrefix)
	if err != nil {
		if !validation.AddValidation(err) {
			return domaintcp.ResultStorageInput{}, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return domaintcp.ResultStorageInput{}, err
	}

	return domaintcp.ResultStorageInput{
		StartedAt:         timing.startedAt,
		FinishedAt:        timing.finishedAt,
		DurationMs:        timing.durationMs,
		Status:            status,
		ConnectDurationMs: connectDurationMs,
		ResolvedIP:        cloneAddr(input.ResolvedIP),
		IPFamily:          metadata.ipFamily,
		ErrorCode:         metadata.errorCode,
		ErrorMessage:      metadata.errorMessage,
	}, nil
}

func normalizeTCPStatus(input, fieldPrefix string) (domaintcp.Status, error) {
	status, err := domaintcp.VNResultStatus(domaintcp.Status(input))
	if err != nil {
		return "", invalidRuntimeField(resultField(fieldPrefix, "status"), err.Error(), input)
	}
	return status, nil
}

func normalizeTCPMetadata(input TCPResultInput, fieldPrefix string) (normalizedTCPMetadata, error) {
	metadata, err := normalizeResultMetadata(input.IPFamily, input.ErrorCode, input.ErrorMessage, fieldPrefix, normalizeOptionalTCPText)
	if err != nil {
		return normalizedTCPMetadata{}, err
	}

	return normalizedTCPMetadata(metadata), nil
}

func normalizeOptionalTCPText(input *string, field string) (*string, error) {
	value, err := domaintcp.VNResultOptionalText(input)
	if err != nil {
		return nil, invalidRuntimeField(field, err.Error(), input)
	}
	return value, nil
}
