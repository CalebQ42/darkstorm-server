package darkstorm

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

func generateSalt() (string, error) {
	out := make([]byte, 16)
	_, err := rand.Read(out)
	return base64.RawStdEncoding.EncodeToString(out), err
}

type User struct {
	Perm           map[string]string
	ID             string
	Username       string
	Password       string
	Salt           string
	Email          string
	PasswordChange uint64
}

func NewUser(username, password, email string) (*User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	salt, err := generateSalt()
	if err != nil {
		return nil, err
	}
	out := &User{
		Perm:     make(map[string]string),
		ID:       id.String(),
		Username: username,
		Salt:     salt,
		Email:    email,
	}
	out.Password, err = out.HashPassword(password)
	return out, err
}

func (u *User) HashPassword(password string) (string, error) {
	salt, err := base64.RawStdEncoding.DecodeString(u.Salt)
	if err != nil {
		return "", err
	}
	res := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return base64.RawStdEncoding.EncodeToString(res), nil
}

func (u *User) ValidatePassword(password string) (bool, error) {
	hsh, err := u.HashPassword(password)
	if err != nil {
		return false, err
	}
	return hsh == u.Password, nil
}
