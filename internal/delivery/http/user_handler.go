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

func (h *UserHandler) CreateUser(ctx *ginx.Context) error {
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

func (h *UserHandler) GetUserByID(ctx *ginx.Context) error {
	id := ctx.PathVar()["id"]
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

func (h *UserHandler) GetUsers(ctx *ginx.Context) error {
	limit := 10
	offset := 0

	if limitStr := ctx.Query()["limit"]; limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := ctx.Query()["offset"]; offsetStr != "" {
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

func (h *UserHandler) UpdateUser(ctx *ginx.Context) error {
	id := ctx.PathVar()["id"]
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

func (h *UserHandler) DeleteUser(ctx *ginx.Context) error {
	id := ctx.PathVar()["id"]
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
