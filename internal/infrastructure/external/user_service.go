package external

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.io/xhkzeroone/goframex/internal/domain"
	"github.io/xhkzeroone/goframex/pkg/http/restyx"

	"github.io/xhkzeroone/goframex/pkg/logger/logrusx"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	client *restyx.Client
}

func NewUserService(client *restyx.Client) domain.UserService {
	// Add middleware for logging
	client.Use(func(next restyx.Handler) restyx.Handler {
		return func(req *restyx.Request) error {
			logrusx.Log.Infof("Making external API request: %s %s", req.Method, req.Path)
			err := next(req)
			if err != nil {
				logrusx.Log.Errorf("External API request failed: %v", err)
			} else {
				logrusx.Log.Infof("External API request successful: %s %s", req.Method, req.Path)
			}
			return err
		}
	})

	return &userService{
		client: client,
	}
}

// ExternalUser represents the user structure from external API
type ExternalUser struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Website  string `json:"website"`
	Address  struct {
		Street  string `json:"street"`
		Suite   string `json:"suite"`
		City    string `json:"city"`
		Zipcode string `json:"zipcode"`
	} `json:"address"`
	Company struct {
		Name        string `json:"name"`
		CatchPhrase string `json:"catchPhrase"`
		Bs          string `json:"bs"`
	} `json:"company"`
}

func (s *userService) FetchUserByID(id string) (*domain.User, error) {
	logrusx.Log.Infof("Fetching user from external service: %s", id)

	// Create request to external API
	req := restyx.NewRequest().
		WithContext(context.Background()).
		MethodGet().
		WithPath("/users/{id}").
		AddPathVar("id", id)

	var externalUser ExternalUser
	err := s.client.Exchange(req.Build(), &externalUser)
	if err != nil {
		logrusx.Log.Errorf("Failed to fetch user from external service: %v", err)
		return nil, fmt.Errorf("user not found in external service: %s", id)
	}

	// Validate external user data
	if externalUser.ID == 0 {
		return nil, fmt.Errorf("invalid user data from external service: %s", id)
	}

	// Convert external user to domain user
	user := &domain.User{
		ID:       fmt.Sprintf("%d", externalUser.ID),
		Name:     externalUser.Name,
		Email:    externalUser.Email,
		Phone:    externalUser.Phone,
		Address:  fmt.Sprintf("%s, %s, %s", externalUser.Address.Street, externalUser.Address.City, externalUser.Address.Zipcode),
		IsActive: true,
		// Generate a default password for external users
		Password: "external_user_default_password",
	}

	logrusx.Log.Infof("Successfully fetched user from external service: %s", id)
	return user, nil
}

func (s *userService) FetchUsers(limit, offset int) ([]*domain.User, error) {
	logrusx.Log.Infof("Fetching users from external service: limit=%d, offset=%d", limit, offset)

	// Create request to external API
	req := restyx.NewRequest().
		WithContext(context.Background()).
		MethodGet().
		WithPath("/users").
		AddParam("_limit", fmt.Sprintf("%d", limit)).
		AddParam("_start", fmt.Sprintf("%d", offset))

	var externalUsers []ExternalUser
	err := s.client.Exchange(req.Build(), &externalUsers)
	if err != nil {
		logrusx.Log.Errorf("Failed to fetch users from external service: %v", err)
		return nil, fmt.Errorf("failed to fetch users from external service: %v", err)
	}

	// Convert external users to domain users
	var users []*domain.User
	for _, extUser := range externalUsers {
		if extUser.ID == 0 {
			continue // Skip invalid users
		}
		user := &domain.User{
			ID:       fmt.Sprintf("%d", extUser.ID),
			Name:     extUser.Name,
			Email:    extUser.Email,
			Phone:    extUser.Phone,
			Address:  fmt.Sprintf("%s, %s, %s", extUser.Address.Street, extUser.Address.City, extUser.Address.Zipcode),
			IsActive: true,
			Password: "external_user_default_password",
		}
		users = append(users, user)
	}

	logrusx.Log.Infof("Successfully fetched %d users from external service", len(users))
	return users, nil
}

func (s *userService) IsExternalServiceAvailable() bool {
	// Simple health check to external service
	req := restyx.NewRequest().
		WithContext(context.Background()).
		MethodGet().
		WithPath("/users")

	var users []ExternalUser
	err := s.client.Exchange(req.Build(), &users)
	return err == nil
}

func (s *userService) ValidateUser(user *domain.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if strings.TrimSpace(user.Name) == "" {
		return fmt.Errorf("name is required")
	}

	if len(user.Name) < 2 || len(user.Name) > 50 {
		return fmt.Errorf("name must be between 2 and 50 characters")
	}

	if strings.TrimSpace(user.Email) == "" {
		return fmt.Errorf("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(user.Email) {
		return fmt.Errorf("invalid email format")
	}

	if strings.TrimSpace(user.Password) == "" {
		return fmt.Errorf("password is required")
	}

	if len(user.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}

	if user.Age < 0 || user.Age > 150 {
		return fmt.Errorf("age must be between 0 and 150")
	}

	return nil
}

func (s *userService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logrusx.Log.Errorf("Failed to hash password: %v", err)
		return "", err
	}
	return string(hashedBytes), nil
}

func (s *userService) ComparePassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		logrusx.Log.Errorf("Password comparison failed: %v", err)
		return fmt.Errorf("invalid password")
	}
	return nil
}
