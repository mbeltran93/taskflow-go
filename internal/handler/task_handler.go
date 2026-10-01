package handler

import (
	"net/http"
	"strconv"

	"github.com/mbeltran93/taskflow-go/internal/model"
	"github.com/mbeltran93/taskflow-go/internal/service"
)

type TaskHandler struct {
	tasks *service.TaskService
}

func NewTaskHandler(tasks *service.TaskService) *TaskHandler {
	return &TaskHandler{tasks: tasks}
}

// Create handles POST /api/tasks.
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t, err := h.tasks.Create(r.Context(), service.CreateTaskInput{
		Title: req.Title, Description: req.Description, ProjectID: req.ProjectID,
		AssigneeID: req.AssigneeID, DueDate: req.DueDate,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toTaskResponse(t))
}

// List handles GET /api/tasks?projectId=X&status=Y.
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	var in service.ListTasksInput

	if raw := r.URL.Query().Get("projectId"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "projectId must be a number")
			return
		}
		in.ProjectID = &id
	}

	if raw := r.URL.Query().Get("status"); raw != "" {
		status := model.Status(raw)
		in.Status = &status
	}

	tasks, err := h.tasks.List(r.Context(), in)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	out := make([]taskResponse, 0, len(tasks))
	for i := range tasks {
		out = append(out, toTaskResponse(&tasks[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

// Get handles GET /api/tasks/{id}.
func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}

	t, err := h.tasks.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toTaskResponse(t))
}

// Update handles PUT /api/tasks/{id}.
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}

	var req taskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t, err := h.tasks.Update(r.Context(), id, service.UpdateTaskInput{
		Title: req.Title, Description: req.Description, AssigneeID: req.AssigneeID, DueDate: req.DueDate,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toTaskResponse(t))
}

// UpdateStatus handles PATCH /api/tasks/{id}/status.
func (h *TaskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}

	var req taskStatusRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t, err := h.tasks.UpdateStatus(r.Context(), id, req.Status)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toTaskResponse(t))
}

// Delete handles DELETE /api/tasks/{id}.
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}

	if err := h.tasks.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
