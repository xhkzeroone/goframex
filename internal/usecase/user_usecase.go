package usecase

import (
	"context"
	"fmt"
	"github.io/xhkzeroone/goframex/internal/domain"
	"github.io/xhkzeroone/goframex/pkg/logger/logrusx"
)

type userUsecase struct {
	repo    domain.UserRepository
	service domain.UserService
}

func NewUserUsecase(repo domain.UserRepository, svc domain.UserService) domain.UserUsecase {
	return &userUsecase{repo: repo, service: svc}
}

func (u *userUsecase) CreateUser(ctx context.Context, user *domain.User) error {
	// Validate user data
	if err := u.service.ValidateUser(user); err != nil {
		logrusx.Log.Errorf("User validation failed: %v", err)
		return err
	}

	// Check if email already exists
	existingUser, err := u.repo.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return fmt.Errorf("email already exists: %s", user.Email)
	}

	// Hash password
	hashedPassword, err := u.service.HashPassword(user.Password)
	if err != nil {
		logrusx.Log.Errorf("Failed to hash password: %v", err)
		return err
	}
	user.Password = hashedPassword

	// Create user
	if err := u.repo.Create(ctx, user); err != nil {
		logrusx.Log.Errorf("Failed to create user: %v", err)
		return err
	}

	logrusx.Log.Infof("User created successfully: %s", user.ID)
	return nil
}

func (u *userUsecase) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	// Check if external service is available first
	if u.service.IsExternalServiceAvailable() {
		// Try to get from external service
		externalUser, err := u.service.FetchUserByID(id)
		if err == nil && externalUser != nil {
			logrusx.Log.Infof("User found in external service: %s", id)
			// Cache the user from external service to local database
			if err := u.repo.Create(ctx, externalUser); err != nil {
				logrusx.Log.Warnf("Failed to cache external user to database: %v", err)
			}
			return externalUser, nil
		}
	}

	// If not found in external service or external service unavailable, try local database
	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		logrusx.Log.Errorf("Failed to get user by ID: %v", err)
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		logrusx.Log.Errorf("Failed to get user by email: %v", err)
		return nil, err
	}
	return user, nil
}

func (u *userUsecase) GetAllUsers(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	// Check if external service is available first
	if u.service.IsExternalServiceAvailable() {
		// Try to get from external service
		externalUsers, err := u.service.FetchUsers(limit, offset)
		if err == nil && len(externalUsers) > 0 {
			logrusx.Log.Infof("Found %d users in external service", len(externalUsers))
			// Cache the users from external service to local database
			for _, user := range externalUsers {
				if err := u.repo.Create(ctx, user); err != nil {
					logrusx.Log.Warnf("Failed to cache external user to database: %v", err)
				}
			}
			return externalUsers, nil
		}
	}

	// If not found in external service or external service unavailable, try local database
	users, err := u.repo.GetAll(ctx, limit, offset)
	if err != nil {
		logrusx.Log.Errorf("Failed to get all users: %v", err)
		return nil, err
	}
	return users, nil
}

func (u *userUsecase) UpdateUser(ctx context.Context, user *domain.User) error {
	// Check if user exists
	existingUser, err := u.repo.GetByID(ctx, user.ID)
	if err != nil {
		logrusx.Log.Errorf("User not found for update: %v", err)
		return err
	}

	// Validate user data
	if err := u.service.ValidateUser(user); err != nil {
		logrusx.Log.Errorf("User validation failed: %v", err)
		return err
	}

	// Check if email is being changed and if it already exists
	if user.Email != existingUser.Email {
		userWithEmail, err := u.repo.GetByEmail(ctx, user.Email)
		if err == nil && userWithEmail != nil && userWithEmail.ID != user.ID {
			return fmt.Errorf("email already exists: %s", user.Email)
		}
	}

	// Hash password if it's being changed
	if user.Password != existingUser.Password {
		hashedPassword, err := u.service.HashPassword(user.Password)
		if err != nil {
			logrusx.Log.Errorf("Failed to hash password: %v", err)
			return err
		}
		user.Password = hashedPassword
	}

	// Update user
	if err := u.repo.Update(ctx, user); err != nil {
		logrusx.Log.Errorf("Failed to update user: %v", err)
		return err
	}

	logrusx.Log.Infof("User updated successfully: %s", user.ID)
	return nil
}

func (u *userUsecase) DeleteUser(ctx context.Context, id string) error {
	// Check if user exists
	_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		logrusx.Log.Errorf("User not found for deletion: %v", err)
		return err
	}

	// Delete user
	if err := u.repo.Delete(ctx, id); err != nil {
		logrusx.Log.Errorf("Failed to delete user: %v", err)
		return err
	}

	logrusx.Log.Infof("User deleted successfully: %s", id)
	return nil
}

func (u *userUsecase) GetUserCount(ctx context.Context) (int64, error) {
	count, err := u.repo.Count(ctx)
	if err != nil {
		logrusx.Log.Errorf("Failed to get user count: %v", err)
		return 0, err
	}
	return count, nil
}
