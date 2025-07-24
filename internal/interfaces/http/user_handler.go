package http

import (
	"github.io/xhkzeroone/goframex/internal/domain"
	"net/http"
	"strconv"

	"github.io/xhkzeroone/goframex/pkg/http/ginx"
	"github.io/xhkzeroone/goframex/pkg/logger/logrusx"
)

type UserHandler struct {
	userUsecase domain.UserUsecase
}

func NewUserHandler(userUsecase domain.UserUsecase) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
	}
}

// CreateUserHandler handles POST /users
type CreateUserHandler struct {
	userUsecase domain.UserUsecase
}

func NewCreateUserHandler(userUsecase domain.UserUsecase) *CreateUserHandler {
	return &CreateUserHandler{userUsecase: userUsecase}
}

func (h *CreateUserHandler) Handle(ctx *ginx.Context, body []byte, headers, query, path map[string]string) error {
	var req CreateUserRequest
	if err := ctx.Bind(&req); err != nil {
		logrusx.Log.Errorf("Failed to bind create user request: %v", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return nil
	}

	user := req.ToDomain()
	if err := h.userUsecase.CreateUser(ctx.Request.Context(), user); err != nil {
		logrusx.Log.Errorf("Failed to create user: %v", err)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create user",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return nil
	}

	ctx.JSON(http.StatusCreated, SuccessResponse{
		Message: "User created successfully",
		Data:    NewUserResponse(user),
	})
	return nil
}

// GetUserByIDHandler handles GET /users/:id
type GetUserByIDHandler struct {
	userUsecase domain.UserUsecase
}

func NewGetUserByIDHandler(userUsecase domain.UserUsecase) *GetUserByIDHandler {
	return &GetUserByIDHandler{userUsecase: userUsecase}
}

func (h *GetUserByIDHandler) Handle(ctx *ginx.Context, headers, query, path map[string]string) error {
	id := path["id"]
	if id == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "User ID is required",
			Message: "User ID parameter is missing",
			Code:    http.StatusBadRequest,
		})
		return nil
	}

	user, err := h.userUsecase.GetUserByID(ctx.Request.Context(), id)
	if err != nil {
		logrusx.Log.Errorf("Failed to get user by ID: %v", err)
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "User not found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return nil
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Message: "User retrieved successfully",
		Data:    NewUserResponse(user),
	})
	return nil
}

// GetUsersHandler handles GET /users
type GetUsersHandler struct {
	userUsecase domain.UserUsecase
}

func NewGetUsersHandler(userUsecase domain.UserUsecase) *GetUsersHandler {
	return &GetUsersHandler{userUsecase: userUsecase}
}

func (h *GetUsersHandler) Handle(ctx *ginx.Context, headers, query, path map[string]string) error {
	limit := 10
	offset := 0

	if limitStr := query["limit"]; limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := query["offset"]; offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	users, err := h.userUsecase.GetAllUsers(ctx.Request.Context(), limit, offset)
	if err != nil {
		logrusx.Log.Errorf("Failed to get users: %v", err)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get users",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return nil
	}

	total, err := h.userUsecase.GetUserCount(ctx.Request.Context())
	if err != nil {
		logrusx.Log.Errorf("Failed to get user count: %v", err)
		// Continue without total count
		total = int64(len(users))
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Message: "Users retrieved successfully",
		Data:    NewUsersResponse(users, total, limit, offset),
	})
	return nil
}

// UpdateUserHandler handles PUT /users/:id
type UpdateUserHandler struct {
	userUsecase domain.UserUsecase
}

func NewUpdateUserHandler(userUsecase domain.UserUsecase) *UpdateUserHandler {
	return &UpdateUserHandler{userUsecase: userUsecase}
}

func (h *UpdateUserHandler) Handle(ctx *ginx.Context, body []byte, headers, query, path map[string]string) error {
	id := path["id"]
	if id == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "User ID is required",
			Message: "User ID parameter is missing",
			Code:    http.StatusBadRequest,
		})
		return nil
	}

	var req UpdateUserRequest
	if err := ctx.Bind(&req); err != nil {
		logrusx.Log.Errorf("Failed to bind update user request: %v", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return nil
	}

	user := req.ToDomain(id)
	if err := h.userUsecase.UpdateUser(ctx.Request.Context(), user); err != nil {
		logrusx.Log.Errorf("Failed to update user: %v", err)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update user",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return nil
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Message: "User updated successfully",
		Data:    NewUserResponse(user),
	})
	return nil
}

// DeleteUserHandler handles DELETE /users/:id
type DeleteUserHandler struct {
	userUsecase domain.UserUsecase
}

func NewDeleteUserHandler(userUsecase domain.UserUsecase) *DeleteUserHandler {
	return &DeleteUserHandler{userUsecase: userUsecase}
}

func (h *DeleteUserHandler) Handle(ctx *ginx.Context, body []byte, headers, query, path map[string]string) error {
	id := path["id"]
	if id == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "User ID is required",
			Message: "User ID parameter is missing",
			Code:    http.StatusBadRequest,
		})
		return nil
	}

	if err := h.userUsecase.DeleteUser(ctx.Request.Context(), id); err != nil {
		logrusx.Log.Errorf("Failed to delete user: %v", err)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete user",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return nil
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Message: "User deleted successfully",
	})
	return nil
}
