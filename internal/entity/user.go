package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id           int64
	Uuid         string
	Email        string
	PasswordHash []byte
	FullName     string
	Phone        string
	Role         string
	Birthday     *time.Time
	Created      *time.Time
	Updated      *time.Time
}

var DateLayout string = "2006-01-02"

func NewUser(email string, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return &User{}, err
	}
	return &User{
		Uuid:         uuid.New().String(),
		Email:        email,
		PasswordHash: hash,
	}, nil
}

func (u *User) SetFullName(fullName string) {
	u.FullName = fullName
}

func (u *User) SetPhone(phone *string) {
	if phone == nil {
		u.Phone = ""
		return
	}
	u.Phone = *phone
}

func (u *User) SetBirthday(birthday *string) error {
	if birthday == nil || *birthday == "" {
		u.Birthday = nil
		return nil
	}
	t, err := time.Parse(DateLayout, *birthday)
	if err != nil {
		return errors.New("incorrect date format" + err.Error())
	}
	u.Birthday = &t
	return nil
}
