package label

type ListLabelsInput struct {
	CurrentUserID string
	ProjectRef    string
}

type CreateLabelInput struct {
	CurrentUserID string
	ProjectRef    string
	Key           string
	Value         string
}

type UpdateLabelInput struct {
	CurrentUserID string
	ProjectRef    string
	LabelID       string
	Key           *string
	Value         *string
}

type DeleteLabelInput struct {
	CurrentUserID string
	ProjectRef    string
	LabelID       string
}
