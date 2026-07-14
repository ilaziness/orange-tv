// Package response provides unified API response structures.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/ilaziness/orange-tv/internal/constant"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
)

// Response represents a standard API response.
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Cause   string `json:"cause,omitempty"`
}

// PageData represents paginated data.
type PageData struct {
	List       any   `json:"list"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// isDev checks if the application is running in development mode.
func isDev() bool {
	if config.Cfg == nil {
		return false
	}
	return config.Cfg.App.Env == constant.EnvDev
}

// Success sends a successful response.
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage sends a successful response with custom message.
func SuccessWithMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// SuccessPage sends a successful paginated response.
func SuccessPage(c *gin.Context, list any, total int64, page, pageSize, totalPages int) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    PageData{List: list, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages},
	})
}

// Error sends an error response.
func Error(c *gin.Context, err error) {
	if err == nil {
		Success(c, nil)
		return
	}

	code := 500
	message := "Internal Server Error"
	httpStatus := 500
	var cause string

	if codeErr, ok := errcode.As(err); ok {
		code = codeErr.Code
		message = codeErr.Message
		httpStatus = codeErr.GetHTTPStatus()

		// In development mode, return cause details for debugging
		if isDev() && codeErr.Cause != nil {
			cause = codeErr.Cause.Error()
		}
	}

	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Cause:   cause,
	})
}

// ErrorWithData sends an error response with data.
func ErrorWithData(c *gin.Context, err error, data any) {
	if err == nil {
		Success(c, data)
		return
	}

	code := 500
	message := "Internal Server Error"
	httpStatus := 500
	var cause string

	if codeErr, ok := errcode.As(err); ok {
		code = codeErr.Code
		message = codeErr.Message
		httpStatus = codeErr.GetHTTPStatus()

		// In development mode, return cause details for debugging
		if isDev() && codeErr.Cause != nil {
			cause = codeErr.Cause.Error()
		}
	}

	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    data,
		Cause:   cause,
	})
}
