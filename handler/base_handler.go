package handler

import (
	"math"

	"github.com/dimas292/go-pkg/apperror"
	"github.com/dimas292/go-pkg/model"
	"github.com/dimas292/go-pkg/response"
	"github.com/dimas292/go-pkg/service"
	"github.com/dimas292/go-pkg/validator"
	"github.com/gin-gonic/gin"
)

// BaseHandler provides generic CRUD HTTP handlers for any model.
// No custom code needed — just wire it up and get full CRUD endpoints.
type BaseHandler[T any, PT model.ModelPtr[T]] struct {
	Service *service.BaseService[T, PT]
}

// NewBaseHandler creates a new generic CRUD handler.
func NewBaseHandler[T any, PT model.ModelPtr[T]](svc *service.BaseService[T, PT]) *BaseHandler[T, PT] {
	return &BaseHandler[T, PT]{Service: svc}
}

// Create handles POST / — create a new record.
func (h *BaseHandler[T, PT]) Create(c *gin.Context) {
	entity := PT(new(T))
	if err := c.ShouldBindJSON(entity); err != nil {
		response.HandleError(c, err)
		return
	}

	// Sanitize all string fields
	validator.SanitizeStruct(entity)

	if err := h.Service.Create(entity); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, "created successfully", entity)
}

// FindAll handles GET / — list all records with pagination.
func (h *BaseHandler[T, PT]) FindAll(c *gin.Context) {
	var query response.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.HandleError(c, err)
		return
	}

	entities, total, err := h.Service.FindAll(query.Page, query.PerPage)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	totalPage := int(math.Ceil(float64(total) / float64(query.PerPage)))

	response.Paginated(c, "retrieved successfully", entities, response.Meta{
		Page:      query.Page,
		PerPage:   query.PerPage,
		Total:     total,
		TotalPage: totalPage,
	})
}

// FindByID handles GET /:id — get a single record by UUID.
func (h *BaseHandler[T, PT]) FindByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.HandleError(c, apperror.BadRequest("invalid id"))
		return
	}

	entity, err := h.Service.FindByID(id)
	if err != nil {
		response.HandleError(c, apperror.NotFound("not found"))
		return
	}

	response.Success(c, "retrieved successfully", entity)
}

// Update handles PUT /:id — update an existing record.
func (h *BaseHandler[T, PT]) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.HandleError(c, apperror.BadRequest("invalid id"))
		return
	}

	entity, err := h.Service.FindByID(id)
	if err != nil {
		response.HandleError(c, apperror.NotFound("not found"))
		return
	}

	if err := c.ShouldBindJSON(entity); err != nil {
		response.HandleError(c, err)
		return
	}

	// Sanitize all string fields
	validator.SanitizeStruct(entity)

	if err := h.Service.Update(entity); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, "updated successfully", entity)
}

// Delete handles DELETE /:id — soft-delete a record.
func (h *BaseHandler[T, PT]) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.HandleError(c, apperror.BadRequest("invalid id"))
		return
	}

	if err := h.Service.Delete(id); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, "deleted successfully", nil)
}

// RegisterCRUD registers all 5 CRUD routes on the given router group.
//
//	POST   /          → Create
//	GET    /          → FindAll (paginated)
//	GET    /:id       → FindByID
//	PUT    /:id       → Update
//	DELETE /:id       → Delete
func (h *BaseHandler[T, PT]) RegisterCRUD(rg *gin.RouterGroup) {
	rg.POST("", h.Create)
	rg.GET("", h.FindAll)
	rg.GET("/:id", h.FindByID)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
}
