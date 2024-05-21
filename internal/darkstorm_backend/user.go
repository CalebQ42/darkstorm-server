package darkstorm

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

var (
	ErrPasswordLength = errors.New("password length must be 12-128")
)

func generateSalt() (string, error) {
	out := make([]byte, 16)
	_, err := rand.Read(out)
	return base64.RawStdEncoding.EncodeToString(out), err
}

type User struct {
	Perm           map[string]string `json:"perm" bson:"perm"`
	ID             string            `json:"id" bson:"_id"`
	Username       string            `json:"username" bson:"username"`
	Password       string            `json:"password" bson:"password"`
	Salt           string            `json:"salt" bson:"salt"`
	Email          string            `json:"email" bson:"email"`
	PasswordChange int64             `json:"passwordChange" bson:"passwordChange"`
}

type ReqUser struct {
	Perm     map[string]string
	ID       string
	Username string
}

func (b *Backend) generateJWT(r *ReqUser) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.RegisteredClaims{
		ID:        r.ID,
		Issuer:    "darkstorm.tech",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
	}).SignedString(b.jwtPriv)
}

func (u User) GetID() string {
	return u.ID
}

func (u User) toReqUser() *ReqUser {
	return &ReqUser{
		Perm:     u.Perm,
		ID:       u.ID,
		Username: u.Username,
	}
}

func (u User) HashPassword(password string) (string, error) {
	salt, err := base64.RawStdEncoding.DecodeString(u.Salt)
	if err != nil {
		return "", err
	}
	res := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return base64.RawStdEncoding.EncodeToString(res), nil
}

func (u User) ValidatePassword(password string) (bool, error) {
	hsh, err := u.HashPassword(password)
	if err != nil {
		return false, err
	}
	return hsh == u.Password, nil
}

type createUserRequest struct {
	Username string
	Password string
	Email    string
}

type createUserReturn struct {
	Username string
	Token    string
}

func (b *Backend) CreateUser(w http.ResponseWriter, r *http.Request) {
	//TODO
}

type loginRequest struct {
	Username string
	Password string
}

type loginReturn struct {
	Token   string
	Timeout int
}

func (b *Backend) Login(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.ParseHeader(r)
	if hdr.k == nil || !hdr.k.Perm["user"] || errors.Is(err, ErrApiKeyUnauthorized) {
		ReturnError(w, http.StatusUnauthorized, "invalidKey", "Application not authorized")
		return
	}
	
}
