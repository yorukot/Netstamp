package probe

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func normalizeCreateProbeInput(input CreateProbeInput) (CreateProbeInput, error) {
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		return CreateProbeInput{}, invalidProbeField("projectRef", err.Error(), input.ProjectRef)
	}

	name, err := domainprobe.VNProbeName(input.Name)
	if err != nil {
		return CreateProbeInput{}, invalidProbeField("name", err.Error(), input.Name)
	}

	subdivisionCode, err := domainprobe.VNProbeOptionalSubdivisionCode(input.SubdivisionCode)
	if err != nil {
		return CreateProbeInput{}, invalidProbeField("subdivisionCode", err.Error(), input.SubdivisionCode)
	}

	latitude, longitude, latitudeErr, longitudeErr := domainprobe.VNProbeCoordinates(input.Latitude, input.Longitude)
	if latitudeErr != nil {
		return CreateProbeInput{}, invalidProbeField("latitude", latitudeErr.Error(), input.Latitude)
	}
	if longitudeErr != nil {
		return CreateProbeInput{}, invalidProbeField("longitude", longitudeErr.Error(), input.Longitude)
	}

	labelIDs, err := domainlabel.VNLabelIDs(input.LabelIDs)
	if err != nil {
		return CreateProbeInput{}, invalidProbeField("labelIds", err.Error(), input.LabelIDs)
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	return CreateProbeInput{
		CurrentUserID:   input.CurrentUserID,
		ProjectRef:      projectRef,
		Name:            name,
		Enabled:         &enabled,
		SubdivisionCode: subdivisionCode,
		Latitude:        latitude,
		Longitude:       longitude,
		LabelIDs:        labelIDs,
	}, nil
}

func normalizeTargetProbeInput(input TargetProbeInput) (TargetProbeInput, error) {
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		return TargetProbeInput{}, invalidProbeField("projectRef", err.Error(), input.ProjectRef)
	}
	probeID, err := domainprobe.VNProbeID(input.ProbeID)
	if err != nil {
		return TargetProbeInput{}, invalidProbeField("probeId", err.Error(), input.ProbeID)
	}

	return TargetProbeInput{
		CurrentUserID: input.CurrentUserID,
		ProjectRef:    projectRef,
		ProbeID:       probeID,
	}, nil
}

func normalizeUpdateProbeInput(input UpdateProbeInput) (UpdateProbeInput, error) {
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		return UpdateProbeInput{}, invalidProbeField("projectRef", err.Error(), input.ProjectRef)
	}
	probeID, err := domainprobe.VNProbeID(input.ProbeID)
	if err != nil {
		return UpdateProbeInput{}, invalidProbeField("probeId", err.Error(), input.ProbeID)
	}

	if !hasUpdateProbeChanges(input) {
		return UpdateProbeInput{}, invalidProbeField("", "at least one field must be provided", nil)
	}

	var output UpdateProbeInput

	if input.Name != nil {
		name, err := domainprobe.VNProbeName(*input.Name)
		if err != nil {
			return UpdateProbeInput{}, invalidProbeField("name", err.Error(), input.Name)
		}
		output.Name = &name
	}
	if input.SubdivisionCode != nil {
		subdivisionCode, err := domainprobe.VNProbeOptionalSubdivisionCode(input.SubdivisionCode)
		if err != nil {
			return UpdateProbeInput{}, invalidProbeField("subdivisionCode", err.Error(), input.SubdivisionCode)
		}
		output.SubdivisionCode = subdivisionCode
	}
	if input.Latitude != nil || input.Longitude != nil {
		latitude, longitude, latitudeErr, longitudeErr := domainprobe.VNProbeCoordinates(input.Latitude, input.Longitude)
		if latitudeErr != nil {
			return UpdateProbeInput{}, invalidProbeField("latitude", latitudeErr.Error(), input.Latitude)
		}
		if longitudeErr != nil {
			return UpdateProbeInput{}, invalidProbeField("longitude", longitudeErr.Error(), input.Longitude)
		}
		output.Latitude = latitude
		output.Longitude = longitude
	}
	if input.LabelIDs != nil {
		labelIDs, err := domainlabel.VNLabelIDs(*input.LabelIDs)
		if err != nil {
			return UpdateProbeInput{}, invalidProbeField("labelIds", err.Error(), input.LabelIDs)
		}
		output.LabelIDs = &labelIDs
	}

	return UpdateProbeInput{
		CurrentUserID:   input.CurrentUserID,
		ProjectRef:      projectRef,
		ProbeID:         probeID,
		Name:            output.Name,
		Enabled:         input.Enabled,
		SubdivisionCode: output.SubdivisionCode,
		Latitude:        output.Latitude,
		Longitude:       output.Longitude,
		LabelIDs:        output.LabelIDs,
	}, nil
}

func hasUpdateProbeChanges(input UpdateProbeInput) bool {
	return input.Name != nil ||
		input.Enabled != nil ||
		input.SubdivisionCode != nil ||
		input.Latitude != nil ||
		input.Longitude != nil ||
		input.LabelIDs != nil
}

func invalidProbeField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
