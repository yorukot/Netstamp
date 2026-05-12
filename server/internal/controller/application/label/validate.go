package label

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func normalizeCreateLabelInput(input CreateLabelInput) (CreateLabelInput, error) {
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		return CreateLabelInput{}, invalidLabelField("projectRef", err.Error(), input.ProjectRef)
	}
	key, err := domainlabel.VNLabelKey(input.Key)
	if err != nil {
		return CreateLabelInput{}, invalidLabelField("key", err.Error(), input.Key)
	}
	value, err := domainlabel.VNLabelValue(input.Value)
	if err != nil {
		return CreateLabelInput{}, invalidLabelField("value", err.Error(), input.Value)
	}

	return CreateLabelInput{ProjectRef: projectRef, Key: key, Value: value}, nil
}

func normalizeUpdateLabelInput(input UpdateLabelInput) (UpdateLabelInput, error) {
	if input.Key == nil && input.Value == nil {
		return UpdateLabelInput{}, invalidLabelField("", "at least one field must be provided", nil)
	}

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		return UpdateLabelInput{}, invalidLabelField("projectRef", err.Error(), input.ProjectRef)
	}
	labelID, err := domainlabel.VNLabelID(input.LabelID)
	if err != nil {
		return UpdateLabelInput{}, invalidLabelField("labelId", err.Error(), input.LabelID)
	}

	var keyPtr *string
	if input.Key != nil {
		key, err := domainlabel.VNLabelKey(*input.Key)
		if err != nil {
			return UpdateLabelInput{}, invalidLabelField("key", err.Error(), input.Key)
		}
		keyPtr = &key
	}

	var valuePtr *string
	if input.Value != nil {
		value, err := domainlabel.VNLabelValue(*input.Value)
		if err != nil {
			return UpdateLabelInput{}, invalidLabelField("value", err.Error(), input.Value)
		}
		valuePtr = &value
	}

	return UpdateLabelInput{ProjectRef: projectRef, LabelID: labelID, Key: keyPtr, Value: valuePtr}, nil
}

func invalidLabelField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
