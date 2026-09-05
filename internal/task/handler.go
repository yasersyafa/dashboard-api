package task

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime/types"

	"github.com/google/uuid"

	"github.com/yasersyafa/dashboard-api/internal/gen"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GET /api/tasks
func (h *Handler) ListTasks(c *gin.Context, params gen.ListTasksParams) {
	filter := FilterAll
	if params.Filter != nil {
		filter = Filter(*params.Filter)
	}

	tasks, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list tasks")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toGenTasks(tasks)})
}

// GET /api/tasks/recent
func (h *Handler) ListRecentTasks(c *gin.Context, params gen.ListRecentTasksParams) {
	take := 4
	if params.Take != nil {
		take = *params.Take
	}

	tasks, err := h.service.ListRecent(c.Request.Context(), take)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list recent tasks")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toGenTasks(tasks)})
}

// GET /api/tasks/stats
func (h *Handler) GetTaskStats(c *gin.Context) {
	stats, err := h.service.Stats(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to get stats")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toGenStats(stats)})
}

// GET /api/tasks/:id
func (h *Handler) GetTask(c *gin.Context, id uuid.UUID) {
	t, err := h.service.Get(c.Request.Context(), id)
	if errors.Is(err, ErrNotFound) {
		respondError(c, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to get task")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toGenTask(t)})
}

// POST /api/tasks
func (h *Handler) CreateTask(c *gin.Context) {
	var req gen.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	input := CreateInput{
		Title:   req.Title,
		Notes:   req.Notes,
		DueDate: fromDatePtr(req.DueDate),
	}
	if req.Priority != nil {
		p := Priority(*req.Priority)
		input.Priority = &p
	}

	created, err := h.service.Create(c.Request.Context(), input)
	if isValidationErr(err) {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create task")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": toGenTask(created)})
}

// PATCH /api/tasks/:id
func (h *Handler) UpdateTask(c *gin.Context, id uuid.UUID) {
	var req gen.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	input := UpdateInput{
		Title:   req.Title,
		Notes:   req.Notes,
		DueDate: fromDatePtr(req.DueDate),
		Done:    req.Done,
	}
	if req.Priority != nil {
		p := Priority(*req.Priority)
		input.Priority = &p
	}

	updated, err := h.service.Update(c.Request.Context(), id, input)
	if errors.Is(err, ErrNotFound) {
		respondError(c, http.StatusNotFound, "task not found")
		return
	}
	if isValidationErr(err) {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update task")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toGenTask(updated)})
}

// POST /api/tasks/:id/toggle
func (h *Handler) ToggleTask(c *gin.Context, id uuid.UUID) {
	updated, err := h.service.Toggle(c.Request.Context(), id)
	if errors.Is(err, ErrNotFound) {
		respondError(c, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to toggle task")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toGenTask(updated)})
}

// DELETE /api/tasks/:id
func (h *Handler) DeleteTask(c *gin.Context, id uuid.UUID) {
	err := h.service.Delete(c.Request.Context(), id)
	if errors.Is(err, ErrNotFound) {
		respondError(c, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete task")
		return
	}

	c.Status(http.StatusNoContent)
}

// DELETE /api/tasks/completed
func (h *Handler) ClearCompletedTasks(c *gin.Context) {
	count, err := h.service.DeleteCompleted(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to clear completed tasks")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"deletedCount": count}})
}

// helpers

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

func isValidationErr(err error) bool {
	switch {
	case errors.Is(err, ErrTitleRequired),
		errors.Is(err, ErrTitleTooLong),
		errors.Is(err, ErrNotesTooLong),
		errors.Is(err, ErrInvalidPriority):
		return true
	default:
		return false
	}
}

func toGenTask(t Task) gen.Task {
	return gen.Task{
		Id: t.ID,
		Title: t.Title,
		Notes: t.Notes,
		Done: t.Done,
		Priority: gen.TaskPriority(t.Priority),
		DueDate: toDatePtr(t.DueDate),
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func toGenTasks(tasks []Task) []gen.Task {
	result := make([]gen.Task, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, toGenTask(t))
	}
	return result
}

func toGenStats(s Stats) gen.TaskStats {
	return gen.TaskStats{
		Total: intPtr(s.Total),
		Done: intPtr(s.Done),
		Active: intPtr(s.Active),
		Overdue: intPtr(s.Overdue),
		PercentDone: float32Ptr(float32(s.PercentDone)),
	}
}

func toDatePtr(t *time.Time) *types.Date {
	if t == nil {
		return nil
	}
	return &types.Date{Time: *t}
}

func fromDatePtr(d *types.Date) *time.Time {
	if d == nil {
		return nil
	}
	t := d.Time
	return &t
}

func intPtr(i int) *int { return &i }
func float32Ptr(f float32) *float32 { return &f }