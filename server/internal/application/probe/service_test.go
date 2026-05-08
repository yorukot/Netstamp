package probe

import (
	"context"
	"errors"
	"testing"
	"time"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

const (
	testProjectID = "22222222-2222-2222-2222-222222222222"
	testLabelID   = "33333333-3333-3333-3333-333333333333"
)

func TestCreateProbeCreatesProbeWithSecretAndNormalizedInput(t *testing.T) {
	enabled := false
	city := " JP-13 "
	latitude := 35.6762
	longitude := 139.6503
	recorder := &recordingProbeEventRecorder{}
	repo := &fakeProbeRepository{}
	projectAccess := &fakeProjectAccess{}
	labelAccess := &fakeLabelAccess{}
	service := NewService(repo, projectAccess, labelAccess, fakeSecretGenerator{
		plaintext: "plain-secret",
		hash:      "secret-hash",
	}, recorder)

	output, err := service.CreateProbe(context.Background(), CreateProbeInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		Name:          " tokyo-vps-1 ",
		Enabled:       &enabled,
		City:          &city,
		Latitude:      &latitude,
		Longitude:     &longitude,
		LabelIDs:      []string{testLabelID, testLabelID},
	})
	if err != nil {
		t.Fatalf("create probe: %v", err)
	}

	if output.Secret != "plain-secret" {
		t.Fatalf("expected plaintext secret, got %q", output.Secret)
	}
	if projectAccess.gotProjectRef != "engineering" {
		t.Fatalf("expected project ref, got %q", projectAccess.gotProjectRef)
	}
	if projectAccess.gotUserID != "user-1" {
		t.Fatalf("expected current user id, got %q", projectAccess.gotUserID)
	}
	if labelAccess.gotProjectID != testProjectID {
		t.Fatalf("expected label project id, got %q", labelAccess.gotProjectID)
	}
	if len(labelAccess.gotLabelIDs) != 1 || labelAccess.gotLabelIDs[0] != testLabelID {
		t.Fatalf("expected label access ids, got %#v", labelAccess.gotLabelIDs)
	}
	if repo.gotCreateInput.Name != "tokyo-vps-1" {
		t.Fatalf("expected trimmed name, got %q", repo.gotCreateInput.Name)
	}
	if repo.gotCreateInput.Enabled {
		t.Fatalf("expected enabled false")
	}
	if repo.gotCreateInput.City == nil || *repo.gotCreateInput.City != "JP-13" {
		t.Fatalf("expected trimmed city, got %#v", repo.gotCreateInput.City)
	}
	if repo.gotCreateInput.Latitude == nil || *repo.gotCreateInput.Latitude != latitude {
		t.Fatalf("expected latitude, got %#v", repo.gotCreateInput.Latitude)
	}
	if repo.gotCreateInput.Longitude == nil || *repo.gotCreateInput.Longitude != longitude {
		t.Fatalf("expected longitude, got %#v", repo.gotCreateInput.Longitude)
	}
	if len(repo.gotCreateInput.LabelIDs) != 1 || repo.gotCreateInput.LabelIDs[0] != testLabelID {
		t.Fatalf("expected deduplicated label ids, got %#v", repo.gotCreateInput.LabelIDs)
	}
	if repo.gotCreateInput.SecretHash != "secret-hash" {
		t.Fatalf("expected secret hash, got %q", repo.gotCreateInput.SecretHash)
	}
	if len(output.Probe.Labels) != 1 || output.Probe.Labels[0].ID != testLabelID {
		t.Fatalf("expected resolved labels on output, got %#v", output.Probe.Labels)
	}
	assertNoProbeEvents(t, recorder)
}

func TestCreateProbeDefaultsEnabledToTrue(t *testing.T) {
	repo := &fakeProbeRepository{}
	service := NewService(repo, &fakeProjectAccess{}, &fakeLabelAccess{}, fakeSecretGenerator{plaintext: "plain-secret", hash: "secret-hash"}, &recordingProbeEventRecorder{})

	_, err := service.CreateProbe(context.Background(), CreateProbeInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		Name:          "tokyo-vps-1",
	})
	if err != nil {
		t.Fatalf("create probe: %v", err)
	}
	if !repo.gotCreateInput.Enabled {
		t.Fatalf("expected enabled to default to true")
	}
}

func TestCreateProbeRejectsInvalidInput(t *testing.T) {
	latitude := 35.6762
	whitespace := "   "

	tests := []struct {
		name  string
		input CreateProbeInput
	}{
		{
			name:  "empty name",
			input: CreateProbeInput{Name: "   "},
		},
		{
			name:  "whitespace city",
			input: CreateProbeInput{Name: "tokyo-vps-1", City: &whitespace},
		},
		{
			name:  "latitude without longitude",
			input: CreateProbeInput{Name: "tokyo-vps-1", Latitude: &latitude},
		},
		{
			name:  "invalid label id",
			input: CreateProbeInput{Name: "tokyo-vps-1", LabelIDs: []string{"not-a-uuid"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeProbeRepository{}
			projectAccess := &fakeProjectAccess{}
			labelAccess := &fakeLabelAccess{}
			service := NewService(repo, projectAccess, labelAccess, fakeSecretGenerator{plaintext: "plain-secret", hash: "secret-hash"}, &recordingProbeEventRecorder{})

			_, err := service.CreateProbe(context.Background(), tt.input)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected invalid input, got %v", err)
			}
			if projectAccess.gotProjectRef != "" {
				t.Fatalf("expected project access not to be called, got project ref %q", projectAccess.gotProjectRef)
			}
			if labelAccess.gotProjectID != "" {
				t.Fatalf("expected label access not to be called, got project id %q", labelAccess.gotProjectID)
			}
		})
	}
}

func TestCreateProbeRejectsLongitudeWithoutLatitude(t *testing.T) {
	longitude := 139.6503
	repo := &fakeProbeRepository{}
	projectAccess := &fakeProjectAccess{}
	labelAccess := &fakeLabelAccess{}
	service := NewService(repo, projectAccess, labelAccess, fakeSecretGenerator{plaintext: "plain-secret", hash: "secret-hash"}, &recordingProbeEventRecorder{})

	_, err := service.CreateProbe(context.Background(), CreateProbeInput{
		Name:      "tokyo-vps-1",
		Longitude: &longitude,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if projectAccess.gotProjectRef != "" {
		t.Fatalf("expected project access not to be called, got project ref %q", projectAccess.gotProjectRef)
	}
	if labelAccess.gotProjectID != "" {
		t.Fatalf("expected label access not to be called, got project id %q", labelAccess.gotProjectID)
	}
}

func TestCreateProbeRejectsInaccessibleProject(t *testing.T) {
	repo := &fakeProbeRepository{}
	service := NewService(repo, &fakeProjectAccess{err: ErrProjectNotFound}, &fakeLabelAccess{}, fakeSecretGenerator{plaintext: "plain-secret", hash: "secret-hash"}, &recordingProbeEventRecorder{})

	_, err := service.CreateProbe(context.Background(), CreateProbeInput{
		CurrentUserID: "user-1",
		ProjectRef:    "missing",
		Name:          "tokyo-vps-1",
	})
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("expected project not found, got %v", err)
	}
	if repo.gotCreateInput.Name != "" {
		t.Fatalf("expected create not to be called, got %#v", repo.gotCreateInput)
	}
}

func TestCreateProbeAllowsOwnerAdminAndEditor(t *testing.T) {
	for _, role := range []domainproject.Role{domainproject.RoleOwner, domainproject.RoleAdmin, domainproject.RoleEditor} {
		t.Run(string(role), func(t *testing.T) {
			repo := &fakeProbeRepository{}
			service := NewService(repo, &fakeProjectAccess{role: role}, &fakeLabelAccess{}, fakeSecretGenerator{plaintext: "plain-secret", hash: "secret-hash"}, &recordingProbeEventRecorder{})

			_, err := service.CreateProbe(context.Background(), CreateProbeInput{
				CurrentUserID: "user-1",
				ProjectRef:    "engineering",
				Name:          "tokyo-vps-1",
			})
			if err != nil {
				t.Fatalf("create probe: %v", err)
			}
			if repo.gotCreateInput.Name != "tokyo-vps-1" {
				t.Fatalf("expected create input, got %#v", repo.gotCreateInput)
			}
		})
	}
}

func TestCreateProbeRejectsViewerBeforeLabelLookupOrSecretGeneration(t *testing.T) {
	recorder := &recordingProbeEventRecorder{}
	repo := &fakeProbeRepository{}
	labelAccess := &fakeLabelAccess{}
	secretGenerator := &recordingSecretGenerator{}
	service := NewService(repo, &fakeProjectAccess{role: domainproject.RoleViewer}, labelAccess, secretGenerator, recorder)

	_, err := service.CreateProbe(context.Background(), CreateProbeInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		Name:          "tokyo-vps-1",
		LabelIDs:      []string{testLabelID},
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if labelAccess.gotProjectID != "" {
		t.Fatalf("expected label access not to be called, got project id %q", labelAccess.gotProjectID)
	}
	if secretGenerator.called {
		t.Fatalf("expected secret generator not to be called")
	}
	if repo.gotCreateInput.Name != "" {
		t.Fatalf("expected create not to be called, got %#v", repo.gotCreateInput)
	}
	assertRecordedProbeEvent(t, recorder, ProbeEvent{
		Name:        ProbeEventCreateFailure,
		Action:      ProbeActionCreate,
		Outcome:     ProbeOutcomeFailure,
		Reason:      ProbeReasonForbidden,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
	})
}

func TestCreateProbeRecordsInvalidInputFailure(t *testing.T) {
	recorder := &recordingProbeEventRecorder{}
	service := NewService(&fakeProbeRepository{}, &fakeProjectAccess{}, &fakeLabelAccess{}, fakeSecretGenerator{
		plaintext: "plain-secret",
		hash:      "secret-hash",
	}, recorder)

	_, err := service.CreateProbe(context.Background(), CreateProbeInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		Name:          "   ",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}

	assertRecordedProbeEvent(t, recorder, ProbeEvent{
		Name:        ProbeEventCreateFailure,
		Action:      ProbeActionCreate,
		Outcome:     ProbeOutcomeFailure,
		Reason:      ProbeReasonInvalidInput,
		ActorUserID: "user-1",
		ProjectRef:  "engineering",
	})
}

func TestCreateProbeRecordsProjectNotFoundFailure(t *testing.T) {
	recorder := &recordingProbeEventRecorder{}
	service := NewService(
		&fakeProbeRepository{},
		&fakeProjectAccess{err: ErrProjectNotFound},
		&fakeLabelAccess{},
		fakeSecretGenerator{plaintext: "plain-secret", hash: "secret-hash"},
		recorder,
	)

	_, err := service.CreateProbe(context.Background(), CreateProbeInput{
		CurrentUserID: "user-1",
		ProjectRef:    "missing",
		Name:          "tokyo-vps-1",
	})
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("expected project not found, got %v", err)
	}

	assertRecordedProbeEvent(t, recorder, ProbeEvent{
		Name:        ProbeEventCreateFailure,
		Action:      ProbeActionCreate,
		Outcome:     ProbeOutcomeFailure,
		Reason:      ProbeReasonProjectNotFound,
		ActorUserID: "user-1",
		ProjectRef:  "missing",
	})
}

func TestCreateProbeRecordsProjectLookupFailure(t *testing.T) {
	recorder := &recordingProbeEventRecorder{}
	lookupErr := errors.New("lookup project")
	service := NewService(
		&fakeProbeRepository{},
		&fakeProjectAccess{err: lookupErr},
		&fakeLabelAccess{},
		fakeSecretGenerator{plaintext: "plain-secret", hash: "secret-hash"},
		recorder,
	)

	_, err := service.CreateProbe(context.Background(), CreateProbeInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		Name:          "tokyo-vps-1",
	})
	if !errors.Is(err, lookupErr) {
		t.Fatalf("expected lookup error, got %v", err)
	}

	assertRecordedProbeEvent(t, recorder, ProbeEvent{
		Name:        ProbeEventCreateFailure,
		Action:      ProbeActionCreate,
		Outcome:     ProbeOutcomeFailure,
		Reason:      ProbeReasonProjectLookupFailed,
		ActorUserID: "user-1",
		ProjectRef:  "engineering",
		Err:         lookupErr,
	})
}

func TestCreateProbeRecordsRoleLookupFailure(t *testing.T) {
	recorder := &recordingProbeEventRecorder{}
	lookupErr := errors.New("lookup role")
	service := NewService(
		&fakeProbeRepository{},
		&fakeProjectAccess{roleErr: lookupErr},
		&fakeLabelAccess{},
		fakeSecretGenerator{plaintext: "plain-secret", hash: "secret-hash"},
		recorder,
	)

	_, err := service.CreateProbe(context.Background(), CreateProbeInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		Name:          "tokyo-vps-1",
	})
	if !errors.Is(err, lookupErr) {
		t.Fatalf("expected role lookup error, got %v", err)
	}

	assertRecordedProbeEvent(t, recorder, ProbeEvent{
		Name:        ProbeEventCreateFailure,
		Action:      ProbeActionCreate,
		Outcome:     ProbeOutcomeFailure,
		Reason:      ProbeReasonRoleLookupFailed,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		Err:         lookupErr,
	})
}

func TestCreateProbeRecordsLabelLookupFailure(t *testing.T) {
	recorder := &recordingProbeEventRecorder{}
	lookupErr := errors.New("lookup labels")
	service := NewService(
		&fakeProbeRepository{},
		&fakeProjectAccess{},
		&fakeLabelAccess{err: lookupErr},
		fakeSecretGenerator{plaintext: "plain-secret", hash: "secret-hash"},
		recorder,
	)

	_, err := service.CreateProbe(context.Background(), CreateProbeInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		Name:          "tokyo-vps-1",
		LabelIDs:      []string{testLabelID},
	})
	if !errors.Is(err, lookupErr) {
		t.Fatalf("expected label lookup error, got %v", err)
	}

	assertRecordedProbeEvent(t, recorder, ProbeEvent{
		Name:        ProbeEventCreateFailure,
		Action:      ProbeActionCreate,
		Outcome:     ProbeOutcomeFailure,
		Reason:      ProbeReasonLabelLookupFailed,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		Err:         lookupErr,
	})
}

func TestCreateProbeRecordsSecretGenerationFailure(t *testing.T) {
	recorder := &recordingProbeEventRecorder{}
	secretErr := errors.New("generate secret")
	service := NewService(
		&fakeProbeRepository{},
		&fakeProjectAccess{},
		&fakeLabelAccess{},
		fakeSecretGenerator{err: secretErr},
		recorder,
	)

	_, err := service.CreateProbe(context.Background(), CreateProbeInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		Name:          "tokyo-vps-1",
	})
	if !errors.Is(err, secretErr) {
		t.Fatalf("expected secret error, got %v", err)
	}

	assertRecordedProbeEvent(t, recorder, ProbeEvent{
		Name:        ProbeEventCreateFailure,
		Action:      ProbeActionCreate,
		Outcome:     ProbeOutcomeFailure,
		Reason:      ProbeReasonSecretGenerateFailed,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		Err:         secretErr,
	})
}

func TestCreateProbeRecordsMissingLabelFailureWithoutSecrets(t *testing.T) {
	recorder := &recordingProbeEventRecorder{}
	service := NewService(
		&fakeProbeRepository{},
		&fakeProjectAccess{},
		&fakeLabelAccess{err: ErrLabelNotFound},
		fakeSecretGenerator{plaintext: "plain-secret", hash: "secret-hash"},
		recorder,
	)

	_, err := service.CreateProbe(context.Background(), CreateProbeInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		Name:          "tokyo-vps-1",
		LabelIDs:      []string{testLabelID},
	})
	if !errors.Is(err, ErrLabelNotFound) {
		t.Fatalf("expected label not found, got %v", err)
	}

	assertRecordedProbeEvent(t, recorder, ProbeEvent{
		Name:        ProbeEventCreateFailure,
		Action:      ProbeActionCreate,
		Outcome:     ProbeOutcomeFailure,
		Reason:      ProbeReasonLabelNotFound,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
	})
}

func assertRecordedProbeEvent(t *testing.T, recorder *recordingProbeEventRecorder, want ProbeEvent) {
	t.Helper()

	if len(recorder.events) != 1 {
		t.Fatalf("expected one event, got %d: %#v", len(recorder.events), recorder.events)
	}

	got := recorder.events[0]
	if got.Name != want.Name ||
		got.Action != want.Action ||
		got.Outcome != want.Outcome ||
		got.Reason != want.Reason ||
		got.ActorUserID != want.ActorUserID ||
		got.ProjectID != want.ProjectID ||
		got.ProjectRef != want.ProjectRef ||
		got.ProbeID != want.ProbeID ||
		!errors.Is(got.Err, want.Err) {
		t.Fatalf("unexpected event:\n got: %#v\nwant: %#v", got, want)
	}
}

func assertNoProbeEvents(t *testing.T, recorder *recordingProbeEventRecorder) {
	t.Helper()

	if len(recorder.events) != 0 {
		t.Fatalf("expected no events, got %d: %#v", len(recorder.events), recorder.events)
	}
}

type recordingProbeEventRecorder struct {
	events []ProbeEvent
}

func (r *recordingProbeEventRecorder) RecordProbeEvent(_ context.Context, event ProbeEvent) {
	r.events = append(r.events, event)
}

type fakeProbeRepository struct {
	gotCreateInput domainprobe.CreateProbeStorageInput
	createErr      error
}

func (r *fakeProbeRepository) CreateProbe(_ context.Context, input domainprobe.CreateProbeStorageInput) (domainprobe.Probe, error) {
	r.gotCreateInput = input
	if r.createErr != nil {
		return domainprobe.Probe{}, r.createErr
	}
	return domainprobe.Probe{
		ID:        "probe-1",
		ProjectID: input.ProjectID,
		Name:      input.Name,
		Enabled:   input.Enabled,
		City:      input.City,
		Latitude:  input.Latitude,
		Longitude: input.Longitude,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

type fakeProjectAccess struct {
	gotProjectRef    string
	gotUserID        string
	gotRoleProjectID string
	gotRoleUserID    string
	role             domainproject.Role
	err              error
	roleErr          error
}

func (r *fakeProjectAccess) GetProjectForUser(_ context.Context, projectRef, userID string) (domainproject.Project, error) {
	r.gotProjectRef = projectRef
	r.gotUserID = userID
	if r.err != nil {
		return domainproject.Project{}, r.err
	}
	return domainproject.Project{ID: testProjectID, Slug: "engineering"}, nil
}

func (r *fakeProjectAccess) GetMemberRole(_ context.Context, projectID, userID string) (domainproject.Role, error) {
	r.gotRoleProjectID = projectID
	r.gotRoleUserID = userID
	if r.roleErr != nil {
		return "", r.roleErr
	}
	if r.role != "" {
		return r.role, nil
	}
	return domainproject.RoleOwner, nil
}

type fakeLabelAccess struct {
	gotProjectID string
	gotLabelIDs  []string
	err          error
}

func (r *fakeLabelAccess) GetActiveLabelsByIDsForProject(_ context.Context, projectID string, labelIDs []string) ([]domainlabel.Label, error) {
	r.gotProjectID = projectID
	r.gotLabelIDs = append([]string(nil), labelIDs...)
	if r.err != nil {
		return nil, r.err
	}
	if len(labelIDs) == 0 {
		return []domainlabel.Label{}, nil
	}
	return []domainlabel.Label{{
		ID:        testLabelID,
		ProjectID: projectID,
		Key:       "region",
		Value:     "tokyo",
	}}, nil
}

type fakeSecretGenerator struct {
	plaintext string
	hash      string
	err       error
}

func (g fakeSecretGenerator) GenerateProbeSecret() (string, string, error) {
	if g.err != nil {
		return "", "", g.err
	}
	return g.plaintext, g.hash, nil
}

type recordingSecretGenerator struct {
	called bool
}

func (g *recordingSecretGenerator) GenerateProbeSecret() (string, string, error) {
	g.called = true
	return "plain-secret", "secret-hash", nil
}
