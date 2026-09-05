package finance

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yasersyafa/dashboard-api/internal/gen"
	"github.com/yasersyafa/dashboard-api/pkg/httpx"
)

type CategoryHandler struct {
	service *CategoryService
}

func NewCategoryHandler(service *CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

// GET /api/finance/categories
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	categories, err := h.service.List(c)
	if err != nil {
		httpx.RespondError(c, http.StatusInternalServerError, "failed to list categories")
		return
	}
	httpx.RespondData(c, http.StatusOK, toGenCategories(categories))
}

// GET /api/finance/categories/usage
func (h *CategoryHandler) ListCategoryUsage(c *gin.Context) {
	usages, err := h.service.ListUsage(c)
	if err != nil {
		httpx.RespondError(c, http.StatusInternalServerError, "failed to list category usages")
		return
	}
	httpx.RespondData(c, http.StatusOK, toGenCategoryUsages(usages))
}

// POST /api/finance/categories
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req gen.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	var budget *int64
	if req.Budget != nil {
		b := int64(*req.Budget)
		budget = &b
	}

	created, err := h.service.Create(c.Request.Context(), CreateCategoryInput{
		Name: req.Name,
		Budget: budget,
	})

	if isCategoryValidationErr(err) {
		httpx.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrCategoryNameTaken) {
		httpx.RespondError(c, http.StatusConflict, err.Error())
		return
	}

	if err != nil {
		httpx.RespondError(c, http.StatusInternalServerError, "failed to create category")
		return
	}

	httpx.RespondData(c, http.StatusCreated, toGenCategory(created))
}

// PATCH /api/finance/categories/:id
func (h *CategoryHandler) UpdateCategory(c *gin.Context, id uuid.UUID) {
	var req gen.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	var budget *int64
	if req.Budget != nil {
		b := int64(*req.Budget)
		budget = &b
	}

	updated, err := h.service.Update(c.Request.Context(), id, UpdateCategoryInput{
		Name: req.Name,
		Budget: budget,
	})

	if errors.Is(err, ErrCategoryNotFound) {
		httpx.RespondError(c, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrCategoryNameTaken) {
		httpx.RespondError(c, http.StatusConflict, err.Error())
		return
	}

	if isCategoryValidationErr(err) {
		httpx.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		httpx.RespondError(c, http.StatusInternalServerError, "failed to update category")
		return
	}

	httpx.RespondData(c, http.StatusOK, toGenCategory(updated))
}

// DELETE /api/finance/categories/:id
func (h *CategoryHandler) DeleteCategory(c *gin.Context, id uuid.UUID) {
	err := h.service.Delete(c.Request.Context(), id)
	if errors.Is(err, ErrCategoryNotFound) {
		httpx.RespondError(c, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrCategoryHasTransactions) {
		httpx.RespondError(c, http.StatusConflict, err.Error())
		return
	}

	if err != nil {
		httpx.RespondError(c, http.StatusInternalServerError, "failed to delete category")
		return
	}

	c.Status(http.StatusNoContent)
}

// helper functions
func isCategoryValidationErr(err error) bool {
	switch {
	case errors.Is(err, ErrCategoryNameRequired),
		errors.Is(err, ErrCategoryNameTooLong),
		errors.Is(err, ErrCategoryBudgetInvalid):
		return true
	default:
		return false
	}
}

func toGenCategory(c Category) gen.Category {
	var budget *int
	if c.Budget != nil {
		b := int(*c.Budget)
		budget = &b
	}

	return gen.Category{
		Id: c.ID,
		Name: c.Name,
		NameKey: c.NameKey,
		Budget: budget,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func toGenCategories(categories []Category) []gen.Category {
	result := make([]gen.Category, 0, len(categories))
	for _, c := range categories {
		result = append(result, toGenCategory(c))
	}
	return result
}

func toGenCategoryUsage(u CategoryUsage) gen.CategoryUsage {
	var budget *int
	if u.Budget != nil {
		b := int(*u.Budget)
		budget = &b
	}

	return gen.CategoryUsage{
		Id: u.ID,
		Name: u.Name,
		NameKey: u.NameKey,
		Budget: budget,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Spent: int(u.Spent),
		PercentUsed: u.PercentUsed,
		OverBudget: u.OverBudget,
	}
}

func toGenCategoryUsages(usages []CategoryUsage) []gen.CategoryUsage {
	result := make([]gen.CategoryUsage, 0, len(usages))
	for _, u := range usages {
		result = append(result, toGenCategoryUsage(u))
	}

	return result
}

