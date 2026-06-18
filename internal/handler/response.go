package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ctm/pkg/i18n"
)

func isErr(err, target error) bool { return errors.Is(err, target) }

type successResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data"`
}

type paginatedResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
	Meta    Meta `json:"meta"`
}

type errorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, successResponse{Success: true, Data: data})
}

// OKMsg returns 200 with a localized message translated by i18n key.
func OKMsg(c *gin.Context, data any, key string) {
	c.JSON(http.StatusOK, successResponse{Success: true, Message: i18n.T(c, key), Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, successResponse{Success: true, Data: data})
}

// CreatedMsg returns 201 with a localized message.
func CreatedMsg(c *gin.Context, data any, key string) {
	c.JSON(http.StatusCreated, successResponse{Success: true, Message: i18n.T(c, key), Data: data})
}

func Paginated(c *gin.Context, data any, total, page, perPage int) {
	c.JSON(http.StatusOK, paginatedResponse{
		Success: true,
		Data:    data,
		Meta:    Meta{Total: total, Page: page, PerPage: perPage},
	})
}

// Fail отправляет ошибку с явным сообщением.
func Fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, errorResponse{
		Success: false,
		Error:   ErrorBody{Code: code, Message: message},
	})
}

// FailCode отправляет ошибку с автоматическим переводом сообщения.
func FailCode(c *gin.Context, status int, code string) {
	Fail(c, status, code, i18n.T(c, code))
}

func BadRequest(c *gin.Context, code, message string) {
	Fail(c, http.StatusBadRequest, code, message)
}

func Unauthorized(c *gin.Context) {
	FailCode(c, http.StatusUnauthorized, "UNAUTHORIZED")
}

func Forbidden(c *gin.Context) {
	FailCode(c, http.StatusForbidden, "FORBIDDEN")
}

func NotFound(c *gin.Context) {
	FailCode(c, http.StatusNotFound, "NOT_FOUND")
}

func InternalError(c *gin.Context) {
	FailCode(c, http.StatusInternalServerError, "INTERNAL_ERROR")
}

// parsePagination reads page/per_page query params with safe defaults.
func parsePagination(c *gin.Context) (page, perPage int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ = strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return
}
