package domain

import (
	"context"
	"time"
)

type User struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"`
	Age       int       `json:"age"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetAll(ctx context.Context, limit, offset int) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}

type UserService interface {
	FetchUserByID(id string) (*User, error)
	FetchUsers(limit, offset int) ([]*User, error)
	IsExternalServiceAvailable() bool
	ValidateUser(user *User) error
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
}

type UserUsecase interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetAllUsers(ctx context.Context, limit, offset int) ([]*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error
	GetUserCount(ctx context.Context) (int64, error)
}
