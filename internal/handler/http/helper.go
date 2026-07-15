package http

import (
	"github.com/gin-gonic/gin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/response"
	"github.com/ilaziness/orange-tv/internal/validator"
)

// BindAndValidate 绑定并验证请求参数
// 使用gin的ShouldBindJSON/ShouldBindQuery/ShouldBindUri等绑定方法
// 绑定后自动调用validator.Validate进行验证
// 如果绑定或验证失败，自动返回错误响应并返回false
func BindAndValidate(c *gin.Context, obj any) bool {
	if err := c.ShouldBind(obj); err != nil {
		response.Error(c, errcode.Wrap(errcode.ParamError, err))
		return false
	}

	if err := validator.Validate(obj); err != nil {
		response.Error(c, errcode.Wrap(errcode.ParamError, err))
		return false
	}

	return true
}

// BindJSON 绑定JSON请求体
func BindJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		response.Error(c, errcode.Wrap(errcode.ParamError, err))
		return false
	}
	return true
}

// BindQuery 绑定查询参数，并执行 validate 标签校验。
func BindQuery(c *gin.Context, obj any) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		response.Error(c, errcode.Wrap(errcode.ParamError, err))
		return false
	}
	if err := validator.Validate(obj); err != nil {
		response.Error(c, errcode.Wrap(errcode.ParamError, err))
		return false
	}
	return true
}

// BindURI 绑定URI参数
func BindURI(c *gin.Context, obj any) bool {
	if err := c.ShouldBindUri(obj); err != nil {
		response.Error(c, errcode.Wrap(errcode.ParamError, err))
		return false
	}
	return true
}

// HandleServiceError 处理service层错误
func HandleServiceError(c *gin.Context, err error) {
	if _, ok := errcode.As(err); ok {
		response.Error(c, err)
		return
	}
	response.Error(c, errcode.Wrap(errcode.InternalError, err))
}
