// Package http provides HTTP handlers for user-related operations.
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/response"
	"github.com/ilaziness/orange-tv/internal/service"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new UserHandler instance.
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetUser retrieves a user by ID.
// @Summary 获取用户信息
// @Description 根据 ID 获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/client/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	var req dto.GetUserRequest
	if !BindURI(c, &req) {
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), req.ID)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, toUserResponse(user))
}

// CreateUser creates a new user.
// @Summary 创建用户
// @Description 创建新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "创建用户请求"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/client/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if !BindAndValidate(c, &req) {
		return
	}

	resp, err := h.userService.Create(c.Request.Context(), &req)
	if err != nil {
		if codeErr, ok := errcode.As(err); ok && codeErr.Code == errcode.UserAlreadyExists.Code {
			response.Error(c, errcode.UserAlreadyExists)
			return
		}
		response.Error(c, errcode.Wrap(errcode.InternalError, err))
		return
	}

	response.Success(c, resp)
}

// UpdateUser updates an existing user.
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Param request body dto.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/client/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req dto.UpdateUserRequest
	if !BindURI(c, &req) {
		return
	}
	if !BindAndValidate(c, &req) {
		return
	}

	resp, err := h.userService.Update(c.Request.Context(), req.ID, &req)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, resp)
}

// DeleteUser deletes a user by ID.
// @Summary 删除用户
// @Description 根据 ID 删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/client/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	var req dto.DeleteUserRequest
	if !BindURI(c, &req) {
		return
	}

	if err := h.userService.Delete(c.Request.Context(), req.ID); err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"id": req.ID})
}

// ListUsers retrieves a list of users with pagination.
// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/client/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req dto.ListUsersRequest
	if !BindQuery(c, &req) {
		return
	}

	page := req.GetPage()
	limit := req.GetLimit()
	offset := req.GetOffset()

	users, total, err := h.userService.ListWithCount(c.Request.Context(), limit, offset)
	if err != nil {
		response.Error(c, errcode.Wrap(errcode.InternalError, err))
		return
	}

	userResponses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = toUserResponse(user)
	}

	response.SuccessPage(c, userResponses, int64(total), page, limit, req.GetTotalPages(total))
}

func toUserResponse(user *model.User) *dto.UserResponse {
	if user == nil {
		return nil
	}
	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Phone:     user.Phone,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
