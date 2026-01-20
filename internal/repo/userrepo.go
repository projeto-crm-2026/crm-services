package repo

import (
	"context"

	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"gorm.io/gorm"
)

type UserRepo interface {
	Insert(ctx context.Context, name string, email string, passwordHash string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Insert(ctx context.Context, name string, email string, passwordHash string) (*entity.User, error) {
	user := entity.User{
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User

	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
