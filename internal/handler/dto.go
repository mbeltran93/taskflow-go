package handler

import (
	"time"

	"github.com/mbeltran93/taskflow-go/internal/model"
)

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func toUserResponse(u *model.User) userResponse {
	return userResponse{ID: u.ID, Name: u.Name, Email: u.Email, CreatedAt: u.CreatedAt}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type projectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     int64  `json:"ownerId"`
}

type projectResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     int64     `json:"ownerId"`
	CreatedAt   time.Time `json:"createdAt"`
}

func toProjectResponse(p *model.Project) projectResponse {
	return projectResponse{ID: p.ID, Name: p.Name, Description: p.Description, OwnerID: p.OwnerID, CreatedAt: p.CreatedAt}
}

type taskRequest struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	ProjectID   int64       `json:"projectId"`
	AssigneeID  *int64      `json:"assigneeId,omitempty"`
	DueDate     *model.Date `json:"dueDate,omitempty"`
}

type taskStatusRequest struct {
	Status model.Status `json:"status"`
}

type taskResponse struct {
	ID          int64        `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      model.Status `json:"status"`
	ProjectID   int64        `json:"projectId"`
	AssigneeID  *int64       `json:"assigneeId,omitempty"`
	DueDate     *model.Date  `json:"dueDate,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
}

func toTaskResponse(t *model.Task) taskResponse {
	return taskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		ProjectID:   t.ProjectID,
		AssigneeID:  t.AssigneeID,
		DueDate:     t.DueDate,
		CreatedAt:   t.CreatedAt,
	}
}
