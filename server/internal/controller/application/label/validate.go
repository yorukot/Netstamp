package label

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func normalizeCreateLabelInput(input CreateLabelInput) (CreateLabelInput, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	key, err := domainlabel.VNLabelKey(input.Key)
	if err != nil {
		validation.AddError("key", err, input.Key)
	}
	value, err := domainlabel.VNLabelValue(input.Value)
	if err != nil {
		validation.AddError("value", err, input.Value)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return CreateLabelInput{}, err
	}

	return CreateLabelInput{
		CurrentUserID: input.CurrentUserID,
		ProjectRef:    projectRef,
		Key:           key,
		Value:         value,
	}, nil
}

func normalizeUpdateLabelInput(input UpdateLabelInput) (UpdateLabelInput, error) {
	if input.Key == nil && input.Value == nil {
		return UpdateLabelInput{}, invalidLabelField("", "at least one field must be provided", nil)
	}

	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	labelID, err := domainlabel.VNLabelID(input.LabelID)
	if err != nil {
		validation.AddError("labelId", err, input.LabelID)
	}

	var keyPtr *string
	if input.Key != nil {
		key, err := domainlabel.VNLabelKey(*input.Key)
		if err != nil {
			validation.AddError("key", err, input.Key)
		} else {
			keyPtr = &key
		}
	}

	var valuePtr *string
	if input.Value != nil {
		value, err := domainlabel.VNLabelValue(*input.Value)
		if err != nil {
			validation.AddError("value", err, input.Value)
		} else {
			valuePtr = &value
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return UpdateLabelInput{}, err
	}

	return UpdateLabelInput{
		CurrentUserID: input.CurrentUserID,
		ProjectRef:    projectRef,
		LabelID:       labelID,
		Key:           keyPtr,
		Value:         valuePtr,
	}, nil
}

func normalizeDeleteLabelInput(input DeleteLabelInput) (DeleteLabelInput, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	labelID, err := domainlabel.VNLabelID(input.LabelID)
	if err != nil {
		validation.AddError("labelId", err, input.LabelID)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return DeleteLabelInput{}, err
	}

	return DeleteLabelInput{
		CurrentUserID: input.CurrentUserID,
		ProjectRef:    projectRef,
		LabelID:       labelID,
	}, nil
}

func invalidLabelField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
