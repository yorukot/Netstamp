package alert

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

const (
	testProjectID      = "11111111-1111-1111-1111-111111111111"
	testProjectRef     = "main"
	testUserID         = "22222222-2222-2222-2222-222222222222"
	testRuleID         = "33333333-3333-3333-3333-333333333333"
	testNotificationID = "44444444-4444-4444-4444-444444444444"
	testIncidentID     = "55555555-5555-5555-5555-555555555555"
)

func TestServiceReadMethodsUseProjectAccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(*Service) error
	}{
		{
			name: "list rules",
			run: func(service *Service) error {
				_, err := service.ListRules(context.Background(), ListRulesInput{ProjectInput: testProjectInput()})
				return err
			},
		},
		{
			name: "get rule",
			run: func(service *Service) error {
				_, err := service.GetRule(context.Background(), GetRuleInput{ProjectInput: testProjectInput(), RuleID: testRuleID})
				return err
			},
		},
		{
			name: "list notifications",
			run: func(service *Service) error {
				_, err := service.ListNotifications(context.Background(), ListNotificationsInput{ProjectInput: testProjectInput()})
				return err
			},
		},
		{
			name: "get notification",
			run: func(service *Service) error {
				_, err := service.GetNotification(context.Background(), GetNotificationInput{ProjectInput: testProjectInput(), NotificationID: testNotificationID})
				return err
			},
		},
		{
			name: "list incidents",
			run: func(service *Service) error {
				_, err := service.ListIncidents(context.Background(), ListIncidentsInput{ProjectInput: testProjectInput()})
				return err
			},
		},
		{
			name: "get incident",
			run: func(service *Service) error {
				_, err := service.GetIncident(context.Background(), GetIncidentInput{ProjectInput: testProjectInput(), IncidentID: testIncidentID})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeAlertRepository{}
			access := &fakeAlertProjectAccess{role: domainproject.RoleViewer}
			service := NewService(repo, access, nil, nil)

			if err := tt.run(service); err != nil {
				t.Fatalf("read method returned error: %v", err)
			}
			if access.projectCalls != 1 {
				t.Fatalf("project access calls = %d, want 1", access.projectCalls)
			}
			if access.roleCalls != 0 {
				t.Fatalf("role calls = %d, want 0", access.roleCalls)
			}
			if repo.lastProjectID != testProjectID {
				t.Fatalf("repo project ID = %q, want %q", repo.lastProjectID, testProjectID)
			}
		})
	}
}

func TestServiceRuleWritesRequireManageAlertPermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(*Service) error
	}{
		{
			name: "create rule",
			run: func(service *Service) error {
				_, err := service.CreateRule(context.Background(), validCreateRuleInput())
				return err
			},
		},
		{
			name: "update rule",
			run: func(service *Service) error {
				input := validUpdateRuleInput()
				_, err := service.UpdateRule(context.Background(), input)
				return err
			},
		},
		{
			name: "delete rule",
			run: func(service *Service) error {
				return service.DeleteRule(context.Background(), DeleteRuleInput{ProjectInput: testProjectInput(), RuleID: testRuleID})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeAlertRepository{}
			access := &fakeAlertProjectAccess{role: domainproject.RoleViewer}
			service := NewService(repo, access, nil, nil)

			err := tt.run(service)
			if !errors.Is(err, ErrForbidden) {
				t.Fatalf("error = %v, want ErrForbidden", err)
			}
			if access.roleCalls != 1 {
				t.Fatalf("role calls = %d, want 1", access.roleCalls)
			}
			if repo.writeCalls != 0 {
				t.Fatalf("repo write calls = %d, want 0", repo.writeCalls)
			}
		})
	}
}

func TestServiceRuleWriteAllowsEditorManageAlertPermission(t *testing.T) {
	t.Parallel()

	repo := &fakeAlertRepository{}
	access := &fakeAlertProjectAccess{role: domainproject.RoleEditor}
	service := NewService(repo, access, nil, nil)

	_, err := service.CreateRule(context.Background(), validCreateRuleInput())
	if err != nil {
		t.Fatalf("create rule returned error: %v", err)
	}
	if repo.writeCalls != 1 {
		t.Fatalf("repo write calls = %d, want 1", repo.writeCalls)
	}
}

func TestServiceNotificationWritesOwnerAdminOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		role       domainproject.Role
		run        func(*Service) error
		writeCalls int
		want       error
	}{
		{
			name:       "owner creates",
			role:       domainproject.RoleOwner,
			writeCalls: 1,
			run: func(service *Service) error {
				_, err := service.CreateNotification(context.Background(), validCreateNotificationInput())
				return err
			},
		},
		{
			name:       "admin creates",
			role:       domainproject.RoleAdmin,
			writeCalls: 1,
			run: func(service *Service) error {
				_, err := service.CreateNotification(context.Background(), validCreateNotificationInput())
				return err
			},
		},
		{
			name: "editor create denied",
			role: domainproject.RoleEditor,
			want: ErrForbidden,
			run: func(service *Service) error {
				_, err := service.CreateNotification(context.Background(), validCreateNotificationInput())
				return err
			},
		},
		{
			name: "editor update denied",
			role: domainproject.RoleEditor,
			want: ErrForbidden,
			run: func(service *Service) error {
				input := validUpdateNotificationInput()
				_, err := service.UpdateNotification(context.Background(), input)
				return err
			},
		},
		{
			name: "editor delete denied",
			role: domainproject.RoleEditor,
			want: ErrForbidden,
			run: func(service *Service) error {
				return service.DeleteNotification(context.Background(), DeleteNotificationInput{ProjectInput: testProjectInput(), NotificationID: testNotificationID})
			},
		},
		{
			name: "editor test denied",
			role: domainproject.RoleEditor,
			want: ErrForbidden,
			run: func(service *Service) error {
				_, err := service.TestNotification(context.Background(), TestNotificationInput{ProjectInput: testProjectInput(), NotificationID: testNotificationID})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeAlertRepository{}
			access := &fakeAlertProjectAccess{role: tt.role}
			service := NewService(repo, access, nil, nil)

			err := tt.run(service)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			if repo.writeCalls != tt.writeCalls {
				t.Fatalf("repo write calls = %d, want %d", repo.writeCalls, tt.writeCalls)
			}
		})
	}
}

func TestServiceCreateRuleRejectsUnsupportedTracerouteCheckType(t *testing.T) {
	t.Parallel()

	repo := &fakeAlertRepository{}
	access := &fakeAlertProjectAccess{role: domainproject.RoleAdmin}
	service := NewService(repo, access, nil, nil)
	input := validCreateRuleInput()
	input.CheckType = domaincheck.TypeTraceroute

	_, err := service.CreateRule(context.Background(), input)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
	if repo.writeCalls != 0 {
		t.Fatalf("repo write calls = %d, want 0", repo.writeCalls)
	}
}

func TestServiceCreateRuleDefaultsTriggerAfterSeconds(t *testing.T) {
	t.Parallel()

	repo := &fakeAlertRepository{}
	access := &fakeAlertProjectAccess{role: domainproject.RoleAdmin}
	service := NewService(repo, access, nil, nil)
	input := validCreateRuleInput()
	input.TriggerAfterSeconds = 0

	created, err := service.CreateRule(context.Background(), input)
	if err != nil {
		t.Fatalf("create rule returned error: %v", err)
	}
	if created.TriggerAfterSeconds != domainalert.DefaultTriggerAfterSeconds {
		t.Fatalf("trigger after seconds = %d, want %d", created.TriggerAfterSeconds, domainalert.DefaultTriggerAfterSeconds)
	}
}

func TestServiceCreateRuleRejectsInvalidTriggerAfterSeconds(t *testing.T) {
	t.Parallel()

	for _, value := range []int32{30, 90, 86460} {
		t.Run(strconv.Itoa(int(value)), func(t *testing.T) {
			t.Parallel()

			repo := &fakeAlertRepository{}
			access := &fakeAlertProjectAccess{role: domainproject.RoleAdmin}
			service := NewService(repo, access, nil, nil)
			input := validCreateRuleInput()
			input.TriggerAfterSeconds = value

			_, err := service.CreateRule(context.Background(), input)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
			if repo.writeCalls != 0 {
				t.Fatalf("repo write calls = %d, want 0", repo.writeCalls)
			}
		})
	}
}

func TestServiceTestNotificationUnavailableReturnsStructuredResult(t *testing.T) {
	t.Parallel()

	repo := &fakeAlertRepository{
		notification: domainalert.Notification{
			ID:        testNotificationID,
			ProjectID: testProjectID,
			Name:      "Pager",
			Type:      domainalert.NotificationTypeWebhook,
			Enabled:   true,
			Config:    json.RawMessage(`{"url":"https://example.com/hook"}`),
		},
	}
	access := &fakeAlertProjectAccess{role: domainproject.RoleAdmin}
	service := NewService(repo, access, nil, nil)

	result, err := service.TestNotification(context.Background(), TestNotificationInput{
		ProjectInput:   testProjectInput(),
		NotificationID: testNotificationID,
	})
	if err != nil {
		t.Fatalf("test notification returned error: %v", err)
	}
	if result.Delivered {
		t.Fatal("delivered = true, want false")
	}
	if result.Kind != "notification" {
		t.Fatalf("kind = %q, want notification", result.Kind)
	}
	if result.Code != "tester_unavailable" {
		t.Fatalf("code = %q, want tester_unavailable", result.Code)
	}
	if result.Message == "" {
		t.Fatal("message is empty")
	}
}

func testProjectInput() ProjectInput {
	return ProjectInput{ProjectRef: testProjectRef, CurrentUserID: testUserID}
}

func validCreateRuleInput() CreateRuleInput {
	return CreateRuleInput{
		ProjectInput:        testProjectInput(),
		Name:                "High packet loss",
		Enabled:             true,
		Severity:            domainalert.SeverityWarning,
		CheckType:           domaincheck.TypePing,
		Condition:           validPingCondition(),
		TriggerAfterSeconds: domainalert.DefaultTriggerAfterSeconds,
		CooldownSeconds:     domainalert.DefaultCooldownSeconds,
		NotificationIDs:     nil,
	}
}

func validUpdateRuleInput() UpdateRuleInput {
	input := validCreateRuleInput()
	return UpdateRuleInput{
		ProjectInput:        input.ProjectInput,
		RuleID:              testRuleID,
		Name:                input.Name,
		Description:         input.Description,
		Enabled:             input.Enabled,
		Severity:            input.Severity,
		CheckType:           input.CheckType,
		ProbeID:             input.ProbeID,
		CheckID:             input.CheckID,
		Condition:           input.Condition,
		TriggerAfterSeconds: input.TriggerAfterSeconds,
		CooldownSeconds:     input.CooldownSeconds,
		NotificationIDs:     input.NotificationIDs,
	}
}

func validPingCondition() alertcondition.Condition {
	return alertcondition.Condition{
		Type:          alertcondition.TypeMetricThreshold,
		Metric:        alertcondition.MetricPingLossPercent,
		Operator:      alertcondition.OperatorGT,
		Threshold:     20,
		WindowSeconds: 300,
		MinSamples:    3,
	}
}

func validCreateNotificationInput() CreateNotificationInput {
	return CreateNotificationInput{
		ProjectInput: testProjectInput(),
		Name:         "Webhook",
		Type:         domainalert.NotificationTypeWebhook,
		Enabled:      true,
		Config:       json.RawMessage(`{"url":"https://example.com/hook"}`),
	}
}

func validUpdateNotificationInput() UpdateNotificationInput {
	input := validCreateNotificationInput()
	return UpdateNotificationInput{
		ProjectInput:   input.ProjectInput,
		NotificationID: testNotificationID,
		Name:           input.Name,
		Type:           input.Type,
		Enabled:        input.Enabled,
		Config:         input.Config,
	}
}

type fakeAlertProjectAccess struct {
	role         domainproject.Role
	projectCalls int
	roleCalls    int
}

func (a *fakeAlertProjectAccess) GetProjectForUser(context.Context, string, string) (domainproject.Project, error) {
	a.projectCalls++
	return domainproject.Project{ID: testProjectID, Slug: testProjectRef, Name: "Main", CreatedByUserID: testUserID}, nil
}

func (a *fakeAlertProjectAccess) GetMemberRole(context.Context, string, string) (domainproject.Role, error) {
	a.roleCalls++
	return a.role, nil
}

type fakeAlertRepository struct {
	lastProjectID string
	writeCalls    int
	notification  domainalert.Notification
}

func (r *fakeAlertRepository) ListRules(_ context.Context, projectID string, _ *domainalert.RuleStatus, _ *domaincheck.Type) ([]domainalert.Rule, error) {
	r.lastProjectID = projectID
	return []domainalert.Rule{}, nil
}

func (r *fakeAlertRepository) GetRule(_ context.Context, projectID, ruleID string) (domainalert.Rule, error) {
	r.lastProjectID = projectID
	return domainalert.Rule{ID: ruleID, ProjectID: projectID}, nil
}

func (r *fakeAlertRepository) CreateRule(_ context.Context, input domainalert.Rule) (domainalert.Rule, error) {
	r.writeCalls++
	r.lastProjectID = input.ProjectID
	input.ID = testRuleID
	return input, nil
}

func (r *fakeAlertRepository) UpdateRule(_ context.Context, input domainalert.Rule) (domainalert.Rule, error) {
	r.writeCalls++
	r.lastProjectID = input.ProjectID
	return input, nil
}

func (r *fakeAlertRepository) DeleteRule(_ context.Context, projectID, _ string) error {
	r.writeCalls++
	r.lastProjectID = projectID
	return nil
}

func (r *fakeAlertRepository) ListNotifications(_ context.Context, projectID string, _ *domainalert.NotificationType) ([]domainalert.Notification, error) {
	r.lastProjectID = projectID
	return []domainalert.Notification{}, nil
}

func (r *fakeAlertRepository) GetNotification(_ context.Context, projectID, notificationID string) (domainalert.Notification, error) {
	r.lastProjectID = projectID
	if r.notification.ID != "" {
		return r.notification, nil
	}
	return domainalert.Notification{ID: notificationID, ProjectID: projectID}, nil
}

func (r *fakeAlertRepository) CreateNotification(_ context.Context, input domainalert.Notification) (domainalert.Notification, error) {
	r.writeCalls++
	r.lastProjectID = input.ProjectID
	input.ID = testNotificationID
	return input, nil
}

func (r *fakeAlertRepository) UpdateNotification(_ context.Context, input domainalert.Notification) (domainalert.Notification, error) {
	r.writeCalls++
	r.lastProjectID = input.ProjectID
	return input, nil
}

func (r *fakeAlertRepository) DeleteNotification(_ context.Context, projectID, _ string) error {
	r.writeCalls++
	r.lastProjectID = projectID
	return nil
}

func (r *fakeAlertRepository) ListIncidents(_ context.Context, projectID string, _ *domainalert.IncidentStatus, _ int32) ([]domainalert.Incident, error) {
	r.lastProjectID = projectID
	return []domainalert.Incident{}, nil
}

func (r *fakeAlertRepository) GetIncident(_ context.Context, projectID, incidentID string) (domainalert.Incident, error) {
	r.lastProjectID = projectID
	return domainalert.Incident{ID: incidentID, ProjectID: projectID}, nil
}
