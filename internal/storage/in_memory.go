package storage

import (
	"oauth/internal/entity"
	"sync"
)

type InMemory struct {
	mtx   sync.RWMutex
	users map[string]*entity.User
}

func NewInMemory() *InMemory {
	return &InMemory{
		users: make(map[string]*entity.User),
	}
}

func (s *InMemory) AddUser(user *entity.User) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.users[user.Email]; ok {
		return ErrUserExists
	}
	s.users[user.Email] = user
	return nil
}

func (s *InMemory) GetUser(email string) (*entity.User, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	u, ok := s.users[email]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}
