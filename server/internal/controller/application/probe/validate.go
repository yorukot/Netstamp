package probe

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func normalizeCreateProbeInput(input CreateProbeInput) (CreateProbeInput, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}

	name, err := domainprobe.VNProbeName(input.Name)
	if err != nil {
		validation.AddError("name", err, input.Name)
	}

	subdivisionCode, err := domainprobe.VNProbeOptionalSubdivisionCode(input.SubdivisionCode)
	if err != nil {
		validation.AddError("subdivisionCode", err, input.SubdivisionCode)
	}

	latitude, longitude, latitudeErr, longitudeErr := domainprobe.VNProbeCoordinates(input.Latitude, input.Longitude)
	if latitudeErr != nil {
		validation.AddError("latitude", latitudeErr, input.Latitude)
	}
	if longitudeErr != nil {
		validation.AddError("longitude", longitudeErr, input.Longitude)
	}

	labelIDs, err := domainlabel.VNLabelIDs(input.LabelIDs)
	if err != nil {
		validation.AddError("labelIds", err, input.LabelIDs)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return CreateProbeInput{}, err
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
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	probeID, err := domainprobe.VNProbeID(input.ProbeID)
	if err != nil {
		validation.AddError("probeId", err, input.ProbeID)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return TargetProbeInput{}, err
	}

	return TargetProbeInput{
		CurrentUserID: input.CurrentUserID,
		ProjectRef:    projectRef,
		ProbeID:       probeID,
	}, nil
}

func normalizeUpdateProbeInput(input UpdateProbeInput) (UpdateProbeInput, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	probeID, err := domainprobe.VNProbeID(input.ProbeID)
	if err != nil {
		validation.AddError("probeId", err, input.ProbeID)
	}

	hasChanges := hasUpdateProbeChanges(input)
	if !hasChanges {
		validation.Add("", "at least one field must be provided", nil)
	}
	if validation.HasErrors() && !hasChanges {
		return UpdateProbeInput{}, validation.Err(ErrInvalidInput)
	}

	output := normalizeUpdateProbeFields(input, &validation)
	if err := validation.Err(ErrInvalidInput); err != nil {
		return UpdateProbeInput{}, err
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

func normalizeUpdateProbeFields(input UpdateProbeInput, validation *appvalidation.Collector) UpdateProbeInput {
	var output UpdateProbeInput

	normalizeUpdateProbeName(input.Name, &output, validation)
	normalizeUpdateProbeLocation(input, &output, validation)
	normalizeUpdateProbeLabels(input.LabelIDs, &output, validation)

	return output
}

func normalizeUpdateProbeName(nameInput *string, output *UpdateProbeInput, validation *appvalidation.Collector) {
	if nameInput == nil {
		return
	}
	name, err := domainprobe.VNProbeName(*nameInput)
	if err != nil {
		validation.AddError("name", err, nameInput)
		return
	}
	output.Name = &name
}

func normalizeUpdateProbeLocation(input UpdateProbeInput, output *UpdateProbeInput, validation *appvalidation.Collector) {
	if input.SubdivisionCode != nil {
		subdivisionCode, err := domainprobe.VNProbeOptionalSubdivisionCode(input.SubdivisionCode)
		if err != nil {
			validation.AddError("subdivisionCode", err, input.SubdivisionCode)
		} else {
			output.SubdivisionCode = subdivisionCode
		}
	}
	if input.Latitude == nil && input.Longitude == nil {
		return
	}
	latitude, longitude, latitudeErr, longitudeErr := domainprobe.VNProbeCoordinates(input.Latitude, input.Longitude)
	if latitudeErr != nil {
		validation.AddError("latitude", latitudeErr, input.Latitude)
	}
	if longitudeErr != nil {
		validation.AddError("longitude", longitudeErr, input.Longitude)
	}
	if latitudeErr == nil && longitudeErr == nil {
		output.Latitude = latitude
		output.Longitude = longitude
	}
}

func normalizeUpdateProbeLabels(labelIDsInput *[]string, output *UpdateProbeInput, validation *appvalidation.Collector) {
	if labelIDsInput == nil {
		return
	}
	labelIDs, err := domainlabel.VNLabelIDs(*labelIDsInput)
	if err != nil {
		validation.AddError("labelIds", err, labelIDsInput)
		return
	}
	output.LabelIDs = &labelIDs
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
