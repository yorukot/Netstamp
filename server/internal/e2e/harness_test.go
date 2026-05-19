//go:build integration

package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

	appassignment "github.com/yorukot/netstamp/internal/controller/application/assignment"
	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
	applabel "github.com/yorukot/netstamp/internal/controller/application/label"
	appprobe "github.com/yorukot/netstamp/internal/controller/application/probe"
	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
	appuser "github.com/yorukot/netstamp/internal/controller/application/user"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	pgassignment "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/assignment"
	pgcheck "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/check"
	pglabel "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/label"
	pgping "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/ping"
	pgprobe "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/probe"
	pgproject "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/project"
	pgresult "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/result"
	pgtraceroute "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/traceroute"
	pguser "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/user"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/security"
	httpserver "github.com/yorukot/netstamp/internal/controller/transport/http"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

const testDatabaseURLEnv = "NETSTAMP_TEST_DATABASE_URL"

type apiSuite struct {
	baseURL string
	client  *http.Client
	pool    *pgxpool.Pool
}

func newAPISuite(t *testing.T) *apiSuite {
	t.Helper()

	adminDatabaseURL := strings.TrimSpace(os.Getenv(testDatabaseURLEnv))
	if adminDatabaseURL == "" {
		t.Skipf("set %s to run backend API E2E tests", testDatabaseURLEnv)
	}

	t.Logf("e2e: connecting to admin database from %s", testDatabaseURLEnv)
	adminDB := openAdminDatabase(t, adminDatabaseURL)
	t.Cleanup(func() {
		if err := adminDB.Close(); err != nil {
			t.Errorf("close admin database: %v", err)
		}
	})

	databaseName := "netstamp_e2e_" + randomHex(t, 6)
	t.Logf("e2e: creating temporary database %q", databaseName)
	createDatabase(t, adminDB, databaseName)
	t.Cleanup(func() {
		t.Logf("e2e: dropping temporary database %q", databaseName)
		dropDatabase(t, adminDB, databaseName)
	})

	testDatabaseURL := databaseURLWithName(t, adminDatabaseURL, databaseName)
	t.Logf("e2e: running migrations on %q", databaseName)
	migrateDatabase(t, testDatabaseURL)

	t.Logf("e2e: opening application database pool for %q", databaseName)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := postgres.NewPool(ctx, postgres.PoolConfig{
		ConnectionString: testDatabaseURL,
		MaxConns:         5,
		MinConns:         0,
		MaxConnLifetime:  time.Minute,
		MaxConnIdleTime:  time.Minute,
	})
	if err != nil {
		t.Fatalf("open test database pool: %v", err)
	}
	t.Cleanup(pool.Close)

	t.Log("e2e: starting in-process HTTP server")
	server := httptest.NewServer(newTestRouter(pool))
	t.Cleanup(func() {
		t.Log("e2e: stopping in-process HTTP server")
		server.Close()
	})
	t.Logf("e2e: HTTP server listening at %s", server.URL)

	return &apiSuite{
		baseURL: server.URL,
		client:  server.Client(),
		pool:    pool,
	}
}

func openAdminDatabase(t *testing.T, databaseURL string) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open admin database: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("ping admin database from %s: %v", testDatabaseURLEnv, err)
	}
	return db
}

func createDatabase(t *testing.T, adminDB *sql.DB, name string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := adminDB.ExecContext(ctx, "CREATE DATABASE "+quoteIdentifier(name)); err != nil {
		t.Fatalf("create test database %q: %v", name, err)
	}
}

func dropDatabase(t *testing.T, adminDB *sql.DB, name string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := adminDB.ExecContext(
		ctx,
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()",
		name,
	); err != nil {
		t.Errorf("terminate test database connections for %q: %v", name, err)
	}
	if _, err := adminDB.ExecContext(ctx, "DROP DATABASE IF EXISTS "+quoteIdentifier(name)); err != nil {
		t.Errorf("drop test database %q: %v", name, err)
	}
}

func migrateDatabase(t *testing.T, databaseURL string) {
	t.Helper()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open migration database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close migration database: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping migration database: %v", err)
	}
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}
	if err := goose.Up(db, migrationsDir(t)); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
}

func migrationsDir(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate e2e test file")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "db", "migrations")
}

func databaseURLWithName(t *testing.T, databaseURL, name string) string {
	t.Helper()

	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse %s: %v", testDatabaseURLEnv, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		t.Fatalf("%s must be a postgres URL, got %q", testDatabaseURLEnv, databaseURL)
	}
	parsed.Path = "/" + name
	return parsed.String()
}

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func newTestRouter(pool *pgxpool.Pool) http.Handler {
	userRepo := pguser.NewUserRepository(pool)
	projectRepo := pgproject.NewProjectRepository(pool)
	labelRepo := pglabel.NewLabelRepository(pool)
	checkRepo := pgcheck.NewCheckRepository(pool)
	probeRepo := pgprobe.NewProbeRepository(pool)
	pingRepo := pgping.NewPingRepository(pool)
	tracerouteRepo := pgtraceroute.NewTracerouteRepository(pool)
	resultRepo := pgresult.NewResultRepository(pool)
	assignmentRepo := pgassignment.NewAssignmentRepository(pool)
	tokenIssuer := security.NewJWTIssuer("e2e-jwt-secret", time.Hour)
	events := noopEvents{}
	assignmentSvc := appassignment.NewService(assignmentRepo, projectRepo, events)
	passwordHasher := security.NewArgon2idPasswordHasher(security.Argon2idConfig{
		MemoryKiB:   1024,
		Iterations:  1,
		Parallelism: 1,
	})

	authSvc := appauth.NewService(userRepo, passwordHasher, tokenIssuer, events)

	return httpserver.NewRouter(httpserver.Dependencies{
		Log:               zap.NewNop(),
		APIVersion:        "v1",
		AuthService:       authSvc,
		AuthVerifier:      tokenIssuer,
		UserService:       appuser.NewService(userRepo, passwordHasher, events),
		AssignmentService: assignmentSvc,
		CheckService:      appcheck.NewService(checkRepo, projectRepo, labelRepo, assignmentSvc, events),
		LabelService:      applabel.NewService(labelRepo, projectRepo, events, assignmentSvc),
		ProbeService:      appprobe.NewService(probeRepo, projectRepo, labelRepo, assignmentSvc, security.NewProbeSecretGenerator(), events),
		ProbeRuntime: appproberuntime.NewService(
			probeRepo,
			pingRepo,
			tracerouteRepo,
			security.NewProbeSecretVerifier(),
			events,
		),
		ProjectService: appproject.NewService(projectRepo, userRepo, events),
		ResultService:  appresult.NewService(pingRepo, tracerouteRepo, resultRepo, projectRepo),
		ReadinessCheck: postgres.NewReadinessCheck(pool),
		RequestTimeout: 15 * time.Second,
	})
}

type noopEvents struct{}

func (noopEvents) RecordAuthEvent(context.Context, appauth.AuthEvent)          {}
func (noopEvents) RecordUserEvent(context.Context, appuser.UserEvent)          {}
func (noopEvents) RecordProjectEvent(context.Context, appproject.ProjectEvent) {}
func (noopEvents) RecordLabelEvent(context.Context, applabel.LabelEvent)       {}
func (noopEvents) RecordCheckEvent(context.Context, appcheck.CheckEvent)       {}
func (noopEvents) RecordProbeEvent(context.Context, appprobe.ProbeEvent)       {}
func (noopEvents) RecordAssignmentEvent(context.Context, appassignment.AssignmentEvent) {
}

func (noopEvents) RecordProbeRuntimeEvent(context.Context, appproberuntime.ProbeRuntimeEvent) {
}

func (s *apiSuite) doJSON(t *testing.T, method, path string, body any, headers map[string]string, wantStatus int, out any) *http.Response {
	t.Helper()

	res := s.do(t, method, path, body, headers, wantStatus)
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("close response body: %v", err)
		}
	}()
	if out == nil {
		return res
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		t.Fatalf("decode %s %s response: %v", method, path, err)
	}
	return res
}

func (s *apiSuite) do(t *testing.T, method, path string, body any, headers map[string]string, wantStatus int) *http.Response {
	t.Helper()

	var reader io.Reader = http.NoBody
	if body != nil {
		var encoded bytes.Buffer
		if err := json.NewEncoder(&encoded).Encode(body); err != nil {
			t.Fatalf("encode %s %s request: %v", method, path, err)
		}
		reader = &encoded
	}

	req, err := http.NewRequest(method, s.baseURL+path, reader)
	if err != nil {
		t.Fatalf("create %s %s request: %v", method, path, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	res, err := s.client.Do(req)
	if err != nil {
		t.Fatalf("send %s %s request: %v", method, path, err)
	}
	if res.StatusCode != wantStatus {
		responseBody, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		t.Fatalf("expected %s %s status %d, got %d: %s", method, path, wantStatus, res.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return res
}

func authCookieHeaders(cookie *http.Cookie) map[string]string {
	return map[string]string{"Cookie": cookie.Name + "=" + cookie.Value}
}

func sessionCookieFromResponse(t *testing.T, res *http.Response) *http.Cookie {
	t.Helper()

	for _, cookie := range res.Cookies() {
		if cookie.Name == httpmiddleware.SessionCookieName {
			return cookie
		}
	}
	t.Fatalf("expected %s cookie in response", httpmiddleware.SessionCookieName)
	return nil
}

func probeHeaders(secret string) map[string]string {
	return map[string]string{"Authorization": "Probe " + secret}
}

func randomHex(t *testing.T, size int) string {
	t.Helper()

	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		t.Fatalf("read random bytes: %v", err)
	}
	return hex.EncodeToString(value)
}
