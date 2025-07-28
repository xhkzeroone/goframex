package http

import (
	"github.io/xhkzeroone/goframex/internal/domain"
)

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Age      int    `json:"age"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

func (r *CreateUserRequest) ToDomain() *domain.User {
	return &domain.User{
		Name:     r.Name,
		Email:    r.Email,
		Password: r.Password,
		Age:      r.Age,
		Phone:    r.Phone,
		Address:  r.Address,
		IsActive: true,
	}
}

type UpdateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password"`
	Age      int    `json:"age"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	IsActive bool   `json:"is_active"`
}

func (r *UpdateUserRequest) ToDomain(id string) *domain.User {
	return &domain.User{
		ID:       id,
		Name:     r.Name,
		Email:    r.Email,
		Password: r.Password,
		Age:      r.Age,
		Phone:    r.Phone,
		Address:  r.Address,
		IsActive: r.IsActive,
	}
}

type GetUsersRequest struct {
	Limit  int `form:"limit" binding:"min=1,max=100"`
	Offset int `form:"offset" binding:"min=0"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Age       int    `json:"age"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func NewUserResponse(user *domain.User) *UserResponse {
	return &UserResponse{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Age:      user.Age,
		Phone:    user.Phone,
		Address:  user.Address,
		IsActive: user.IsActive,
	}
}

type UsersResponse struct {
	Users  []*UserResponse `json:"users"`
	Total  int64           `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

func NewUsersResponse(users []*domain.User, total int64, limit, offset int) *UsersResponse {
	userResponses := make([]*UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = NewUserResponse(user)
	}

	return &UsersResponse{
		Users:  userResponses,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}
