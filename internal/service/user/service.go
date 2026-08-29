package user

import (
	"context"

	"delivery-service/internal/domain"
)

type UserRepositori interface {
	Create(ctx context.Context, user *domain.User) error
}

type Service struct {
	repo UserRepositori
}

func New(
	repo UserRepositori,
) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, dto domain.CreateUserDTO) (*domain.User, error) {
	user := domain.User{
		Name:    dto.Name,
		Address: dto.Address,
		Phone:   dto.Phone,
	}

	if err := s.repo.Create(ctx, &user); err != nil {
		return nil, err
	}

	return &user, nil
}
