// Package integration exercises the full stack - HTTP handlers, services and
// repositories - against a real, disposable PostgreSQL instance started with
// Testcontainers-go. Unlike internal/service's unit tests (which use
// in-memory fakes), this is what actually proves the SQL in
// internal/repository/postgres is correct.
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mbeltran93/taskflow-go/internal/auth"
	"github.com/mbeltran93/taskflow-go/internal/db"
	"github.com/mbeltran93/taskflow-go/internal/handler"
	"github.com/mbeltran93/taskflow-go/internal/repository/postgres"
	"github.com/mbeltran93/taskflow-go/internal/service"
)

// newTestServer starts a Postgres container, migrates it, and returns an
// httptest.Server wired exactly like cmd/api/main.go wires the real one.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "taskflow",
			"POSTGRES_PASSWORD": "taskflow",
			"POSTGRES_DB":       "taskflow",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("terminate container: %v", err)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("container port: %v", err)
	}

	databaseURL := fmt.Sprintf("postgres://taskflow:taskflow@%s:%s/taskflow?sslmode=disable", host, port.Port())

	if err := db.Migrate(databaseURL); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	sqlDB, err := sqlx.Connect("pgx", databaseURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	userRepo := postgres.NewUserRepository(sqlDB)
	projectRepo := postgres.NewProjectRepository(sqlDB)
	taskRepo := postgres.NewTaskRepository(sqlDB)

	jwtManager := auth.NewJWTManager("integration-test-secret", time.Hour)

	router := handler.NewRouter(handler.Handlers{
		Auth:     handler.NewAuthHandler(service.NewAuthService(userRepo, jwtManager)),
		Users:    handler.NewUserHandler(service.NewUserService(userRepo)),
		Projects: handler.NewProjectHandler(service.NewProjectService(projectRepo, userRepo)),
		Tasks:    handler.NewTaskHandler(service.NewTaskService(taskRepo, projectRepo, userRepo)),
	}, jwtManager)

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

func doJSON(t *testing.T, method, url, token string, body any) *http.Response {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}

	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func decodeBody[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	var out T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return out
}

// TestFullTaskLifecycle drives the whole contract end to end: register two
// users, log in, create a project, create a task, move it through its
// statuses, list/filter it, and delete it - plus the auth guard on write routes.
func TestFullTaskLifecycle(t *testing.T) {
	srv := newTestServer(t)
	ctx := context.Background()
	_ = ctx

	// 1. Register an owner and an assignee.
	ownerResp := doJSON(t, http.MethodPost, srv.URL+"/api/users", "", map[string]string{
		"name": "Owner", "email": "owner@example.com", "password": "supersecret",
	})
	if ownerResp.StatusCode != http.StatusCreated {
		t.Fatalf("register owner: status = %d", ownerResp.StatusCode)
	}
	owner := decodeBody[struct {
		ID int64 `json:"id"`
	}](t, ownerResp)

	assigneeResp := doJSON(t, http.MethodPost, srv.URL+"/api/users", "", map[string]string{
		"name": "Assignee", "email": "assignee@example.com", "password": "supersecret",
	})
	if assigneeResp.StatusCode != http.StatusCreated {
		t.Fatalf("register assignee: status = %d", assigneeResp.StatusCode)
	}
	assignee := decodeBody[struct {
		ID int64 `json:"id"`
	}](t, assigneeResp)

	// 2. Creating a project without a token is rejected.
	noAuthResp := doJSON(t, http.MethodPost, srv.URL+"/api/projects", "", map[string]any{
		"name": "Launch", "ownerId": owner.ID,
	})
	if noAuthResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("create project without token: status = %d, want 401", noAuthResp.StatusCode)
	}

	// 3. Log in as the owner.
	loginResp := doJSON(t, http.MethodPost, srv.URL+"/api/auth/login", "", map[string]string{
		"email": "owner@example.com", "password": "supersecret",
	})
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login: status = %d", loginResp.StatusCode)
	}
	login := decodeBody[struct {
		Token string `json:"token"`
	}](t, loginResp)
	if login.Token == "" {
		t.Fatalf("login returned an empty token")
	}

	// 4. Create a project with the token.
	projectResp := doJSON(t, http.MethodPost, srv.URL+"/api/projects", login.Token, map[string]any{
		"name": "Launch", "description": "Q4 launch", "ownerId": owner.ID,
	})
	if projectResp.StatusCode != http.StatusCreated {
		t.Fatalf("create project: status = %d", projectResp.StatusCode)
	}
	project := decodeBody[struct {
		ID int64 `json:"id"`
	}](t, projectResp)

	// 5. Create a task under that project, assigned to the second user.
	taskResp := doJSON(t, http.MethodPost, srv.URL+"/api/tasks", login.Token, map[string]any{
		"title": "Write docs", "projectId": project.ID, "assigneeId": assignee.ID, "dueDate": "2026-12-31",
	})
	if taskResp.StatusCode != http.StatusCreated {
		t.Fatalf("create task: status = %d", taskResp.StatusCode)
	}
	task := decodeBody[struct {
		ID      int64  `json:"id"`
		Status  string `json:"status"`
		DueDate string `json:"dueDate"`
	}](t, taskResp)
	if task.Status != "TODO" {
		t.Errorf("new task status = %q, want TODO", task.Status)
	}
	if task.DueDate != "2026-12-31" {
		t.Errorf("task dueDate = %q, want 2026-12-31", task.DueDate)
	}

	// 6. Move it to IN_PROGRESS.
	statusResp := doJSON(t, http.MethodPatch, fmt.Sprintf("%s/api/tasks/%d/status", srv.URL, task.ID), login.Token,
		map[string]string{"status": "IN_PROGRESS"})
	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("update status: status = %d", statusResp.StatusCode)
	}
	updated := decodeBody[struct {
		Status string `json:"status"`
	}](t, statusResp)
	if updated.Status != "IN_PROGRESS" {
		t.Errorf("task status after PATCH = %q, want IN_PROGRESS", updated.Status)
	}

	// 7. Listing with projectId + status filters finds exactly this task.
	listURL := fmt.Sprintf("%s/api/tasks?projectId=%d&status=IN_PROGRESS", srv.URL, project.ID)
	listResp := doJSON(t, http.MethodGet, listURL, "", nil)
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("list tasks: status = %d", listResp.StatusCode)
	}
	tasks := decodeBody[[]struct {
		ID int64 `json:"id"`
	}](t, listResp)
	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Fatalf("filtered list = %+v, want exactly task %d", tasks, task.ID)
	}

	// 8. Fetching a non-existent project 404s.
	missingResp := doJSON(t, http.MethodGet, srv.URL+"/api/projects/999999", "", nil)
	if missingResp.StatusCode != http.StatusNotFound {
		t.Fatalf("get missing project: status = %d, want 404", missingResp.StatusCode)
	}

	// 9. Delete the task, then confirm it is gone.
	deleteResp := doJSON(t, http.MethodDelete, fmt.Sprintf("%s/api/tasks/%d", srv.URL, task.ID), login.Token, nil)
	if deleteResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete task: status = %d, want 204", deleteResp.StatusCode)
	}
	getDeletedResp := doJSON(t, http.MethodGet, fmt.Sprintf("%s/api/tasks/%d", srv.URL, task.ID), "", nil)
	if getDeletedResp.StatusCode != http.StatusNotFound {
		t.Fatalf("get deleted task: status = %d, want 404", getDeletedResp.StatusCode)
	}
}
