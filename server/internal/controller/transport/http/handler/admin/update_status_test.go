package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"
)

func TestGetUpdateStatusWritesPrivateNoStoreResponse(t *testing.T) {
	publishedAt := time.Date(2026, 8, 20, 9, 0, 0, 0, time.UTC)
	latestVersion := "v1.2.3"
	releaseURL := "https://github.com/yorukot/netstamp/releases/tag/v1.2.3"
	service := appadmin.NewService(updateStatusAdminRepository{})
	service.ConfigureUpdateSettings(updateStatusSettingsReader{enabled: true})
	service.ConfigureUpdateStatus(staticUpdateStatusReader{status: appadmin.UpdateStatus{
		CurrentVersion: "0.0.0", LatestVersion: &latestVersion, UpdateAvailable: true,
		ReleaseURL: &releaseURL, PublishedAt: &publishedAt, LastCheckedAt: &publishedAt,
	}})
	handler := &Handler{service: service}
	recorder := httptest.NewRecorder()

	handler.handleGetUpdateStatus(recorder, authenticatedSettingsRequest(http.MethodGet, "/", ""))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Cache-Control"); got != settingsCacheControl {
		t.Fatalf("expected Cache-Control %q, got %q", settingsCacheControl, got)
	}
	var body updateStatusResponseBody
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode update status: %v", err)
	}
	if body.CurrentVersion != "0.0.0" || body.LatestVersion == nil || *body.LatestVersion != latestVersion || !body.UpdateAvailable {
		t.Fatalf("unexpected update status body: %#v", body)
	}
	if body.ReleaseURL == nil || *body.ReleaseURL != releaseURL || body.PublishedAt == nil || *body.PublishedAt != publishedAt {
		t.Fatalf("unexpected release metadata: %#v", body)
	}
}

type staticUpdateStatusReader struct{ status appadmin.UpdateStatus }

func (r staticUpdateStatusReader) ReadUpdateStatus() appadmin.UpdateStatus { return r.status }

type updateStatusSettingsReader struct{ enabled bool }

func (r updateStatusSettingsReader) UpdateCheckEnabled(context.Context) (bool, error) {
	return r.enabled, nil
}

type updateStatusAdminRepository struct{}

func (updateStatusAdminRepository) IsSystemAdmin(context.Context, string) (bool, error) {
	return true, nil
}

func (updateStatusAdminRepository) ListSystemAdmins(context.Context) ([]appadmin.SystemAdmin, error) {
	return nil, nil
}

func (updateStatusAdminRepository) ListManagedUsers(context.Context) ([]appadmin.ManagedUser, error) {
	return nil, nil
}

func (updateStatusAdminRepository) GrantSystemAdminByEmail(context.Context, string) (appadmin.SystemAdmin, error) {
	return appadmin.SystemAdmin{}, nil
}

func (updateStatusAdminRepository) GrantSystemAdminByUserID(context.Context, string) (appadmin.ManagedUser, error) {
	return appadmin.ManagedUser{}, nil
}

func (updateStatusAdminRepository) RevokeSystemAdminIfNotLast(context.Context, string) (appadmin.SystemAdminRevokeResult, error) {
	return appadmin.SystemAdminRevokeResult{}, nil
}

func (updateStatusAdminRepository) CountActiveSystemAdmins(context.Context) (int64, error) {
	return 1, nil
}

func (updateStatusAdminRepository) SetManagedUserDisabledAt(context.Context, string, bool) (appadmin.ManagedUser, error) {
	return appadmin.ManagedUser{}, nil
}

func (updateStatusAdminRepository) SetManagedUserPasswordHash(context.Context, string, string) (appadmin.ManagedUser, error) {
	return appadmin.ManagedUser{}, nil
}

func (updateStatusAdminRepository) CreateSystemSettingAuditEvent(context.Context, string, string, *string) error {
	return nil
}
