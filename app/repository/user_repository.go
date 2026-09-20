package repository

import (
	"context"

	"latihan-fiber/app/model"
)

type UserRepository interface {
	FindAll(ctx context.Context) ([]model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	Create(ctx context.Context, user model.User) (model.User, error)
	Update(ctx context.Context, user model.User) (model.User, error)
	UpdateRole(ctx context.Context, id int, role string) (model.User, error)
	Delete(ctx context.Context, id int) error
}
