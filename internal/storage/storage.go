package storage

import (
	"errors"
	"oauth/internal/entity"
)

var (
	ErrUserExists = errors.New("user already exists")
	ErrNotFound   = errors.New("not found")
)

type Storage interface {
	AddUser(user *entity.User) error
	GetUser(email string) (*entity.User, error)
	GetUserList(limit int64, offset int64) ([]entity.User, int64, error)
	DeleteUser(int64) error
}
