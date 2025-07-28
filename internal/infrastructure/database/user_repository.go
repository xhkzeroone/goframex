package database

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.io/xhkzeroone/goframex/internal/domain"
	"time"

	"github.io/xhkzeroone/goframex/pkg/cache/redisx"
	"github.io/xhkzeroone/goframex/pkg/database/gormx"
	"github.io/xhkzeroone/goframex/pkg/logger/logrusx"
	"gorm.io/gorm"
)

type UserModel struct {
	ID        uuid.UUID `gorm:"primarykey;column:id;type:uuid"`
	PartnerId string    `gorm:"column:partner_id"`
	Total     int       `gorm:"column:total"`
	UserName  string    `gorm:"column:user_name"`
	FirstName string    `gorm:"column:first_name"`
	LastName  string    `gorm:"column:last_name"`
	Email     string    `gorm:"column:email"`
	Status    string    `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (u *UserModel) TableName() string {
	return "user_tbl"
}

func (u *UserModel) BeforeCreate(ctx *gorm.DB) (err error) {
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	return
}

func (u *UserModel) BeforeUpdate(ctx *gorm.DB) (err error) {
	u.UpdatedAt = time.Now()
	return
}

func (u *UserModel) GetTotal() int {
	return u.Total
}

type userRepository struct {
	db    *gormx.Repository[UserModel, uuid.UUID]
	cache *redisx.Redis
}

func NewUserRepository(db *gormx.DataSource, cache *redisx.Redis) domain.UserRepository {
	return &userRepository{
		db:    gormx.NewRepository[UserModel, uuid.UUID](db),
		cache: cache,
	}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	if err := r.db.Create(user).Error; err != nil {
		logrusx.Log.Errorf("Failed to create user: %v", err)
		return err
	}

	// Cache the created user
	cacheKey := fmt.Sprintf("user:%s", user.ID)
	if err := r.cache.SetJSON(ctx, cacheKey, user, 30*time.Minute); err != nil {
		logrusx.Log.Warnf("Failed to cache user: %v", err)
	}

	logrusx.Log.Infof("User created successfully: %s", user.ID)
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:%s", id)
	var user domain.User
	if err := r.cache.GetJSON(ctx, cacheKey, &user); err == nil {
		logrusx.Log.Infof("User retrieved from cache: %s", id)
		return &user, nil
	}

	// If not in cache, get from database
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		logrusx.Log.Errorf("Failed to get user by ID: %v", err)
		return nil, err
	}

	// Cache the user
	if err := r.cache.SetJSON(ctx, cacheKey, &user, 30*time.Minute); err != nil {
		logrusx.Log.Warnf("Failed to cache user: %v", err)
	}

	logrusx.Log.Infof("User retrieved from database: %s", id)
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		logrusx.Log.Errorf("Failed to get user by email: %v", err)
		return nil, err
	}

	logrusx.Log.Infof("User retrieved by email: %s", email)
	return &user, nil
}

func (r *userRepository) GetAll(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	var users []*domain.User
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logrusx.Log.Errorf("Failed to get all users: %v", err)
		return nil, err
	}

	logrusx.Log.Infof("Retrieved %d users", len(users))
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		logrusx.Log.Errorf("Failed to update user: %v", err)
		return err
	}

	// Update cache
	cacheKey := fmt.Sprintf("user:%s", user.ID)
	if err := r.cache.SetJSON(ctx, cacheKey, user, 30*time.Minute); err != nil {
		logrusx.Log.Warnf("Failed to update cache for user: %v", err)
	}

	logrusx.Log.Infof("User updated successfully: %s", user.ID)
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.User{}).Error; err != nil {
		logrusx.Log.Errorf("Failed to delete user: %v", err)
		return err
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("user:%s", id)
	if err := r.cache.Del(ctx, cacheKey).Err(); err != nil {
		logrusx.Log.Warnf("Failed to remove user from cache: %v", err)
	}

	logrusx.Log.Infof("User deleted successfully: %s", id)
	return nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&count).Error; err != nil {
		logrusx.Log.Errorf("Failed to count users: %v", err)
		return 0, err
	}

	return count, nil
}
