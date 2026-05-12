package label

import "testing"

const (
	testCurrentUserID = "11111111-1111-1111-1111-111111111111"
	testLabelID       = "22222222-2222-2222-2222-222222222222"
)

func TestNormalizeCreateLabelInputPreservesCurrentUserID(t *testing.T) {
	input, err := normalizeCreateLabelInput(CreateLabelInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    " project ",
		Key:           " region ",
		Value:         " tokyo ",
	})
	if err != nil {
		t.Fatalf("expected valid input: %v", err)
	}

	if input.CurrentUserID != testCurrentUserID {
		t.Fatalf("expected current user ID to be preserved, got %q", input.CurrentUserID)
	}
}

func TestNormalizeUpdateLabelInputPreservesCurrentUserID(t *testing.T) {
	key := " region "
	input, err := normalizeUpdateLabelInput(UpdateLabelInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    " project ",
		LabelID:       testLabelID,
		Key:           &key,
	})
	if err != nil {
		t.Fatalf("expected valid input: %v", err)
	}

	if input.CurrentUserID != testCurrentUserID {
		t.Fatalf("expected current user ID to be preserved, got %q", input.CurrentUserID)
	}
}
