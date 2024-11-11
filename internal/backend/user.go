package backend

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

type ReqestUser struct {
	Perm     map[string]string
	ID       string
	Username string
}

func (b *Backend) GenerateJWT(r *ReqestUser) (string, error) {
	if b.jwtPriv == nil || b.jwtPub == nil {
		return "", errors.New("user management not enabled")
	}
	return jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.RegisteredClaims{
		ID:        r.ID,
		Issuer:    "darkstorm.tech",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
	}).SignedString(b.jwtPriv)
}

type User struct {
	Perm           map[string]string `json:"perm" bson:"perm"`
	ID             string            `json:"id" bson:"_id"`
	Username       string            `json:"username" bson:"username"`
	Password       string            `json:"password" bson:"password"`
	Salt           string            `json:"salt" bson:"salt"`
	Email          string            `json:"email" bson:"email"`
	Fails          int               `json:"fails" bson:"fails"`
	Timeout        int64             `json:"timeout" bson:"timeout"`
	PasswordChange int64             `json:"passwordChange" bson:"passwordChange"`
}

var (
	ErrLoginTimeout   = errors.New("user is timed out")
	ErrLoginIncorrect = errors.New("username or password is incorrect")
)

// Tries to login with the given username and password.
// If the user exists, but is timed out, the user is still returned.
func (b *Backend) TryLogin(ctx context.Context, username, password string) (User, error) {
	users, err := b.userTable.Find(ctx, map[string]any{"username": username})
	if err == ErrNotFound {
		return User{}, ErrLoginIncorrect
	}
	if len(users) > 0 {
		log.Println("duplicate username detected, fix immediately:", username)
	}
	user := users[0]
	if time.Unix(user.Timeout, 0).After(time.Now()) {
		return user, ErrLoginTimeout
	}
	if valid, _ := user.ValidatePassword(password); !valid {
		upd := map[string]any{"fails": user.Fails + 1}
		if (user.Fails+1)%3 == 0 {
			minutes := 3 ^ (((user.Fails + 1) / 3) - 1)
			upd["timeout"] = time.Now().Add(time.Minute * time.Duration(minutes)).Unix()
			b.userTable.PartUpdate(ctx, user.ID, upd)
			return user, ErrLoginTimeout
		}
		return User{}, ErrLoginIncorrect
	}
	return user, nil
}

func NewUser(username, password, email string) (User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return User{}, err
	}
	salt, err := generateSalt()
	if err != nil {
		return User{}, err
	}
	u := User{
		Perm:     make(map[string]string),
		ID:       id.String(),
		Username: username,
		Salt:     salt,
		Email:    email,
	}
	u.Password, err = u.HashPassword(password)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (u User) GetID() string {
	return u.ID
}

func (u User) toReqUser() *ReqestUser {
	return &ReqestUser{
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
	Username string `json:"username"`
	Token    string `json:"token"`
}

func (b *Backend) createUser(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.VerifyHeader(w, r, "user", false)
	if hdr == nil {
		if err != nil {
			log.Println("request key parsing error:", err)
		}
		return
	}
	defer r.Body.Close()
	var req createUserRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Username == "" || req.Password == "" || req.Email == "" {
		ReturnError(w, http.StatusBadRequest, "invalidBody", "Bad request")
		return
	}
	if len(req.Password) < 12 || len(req.Password) > 128 {
		ReturnError(w, http.StatusUnauthorized, "password", "Invalid password.")
		return
	}
	// TODO: filter offensive words/phrases
	b.userCreateMutex.Lock()
	defer b.userCreateMutex.Unlock()
	matchUsername, err := b.userTable.Find(r.Context(), map[string]any{"username": req.Username})
	if err != nil && !errors.Is(err, ErrNotFound) {
		log.Println("error when checking for username collisions:", err)
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	} else if (err == nil || errors.Is(err, ErrNotFound)) && len(matchUsername) > 0 {
		ReturnError(w, http.StatusUnauthorized, "taken", "Username or email already used")
		return
	}
	matchEmail, err := b.userTable.Find(r.Context(), map[string]any{"email": req.Email})
	if err != nil && !errors.Is(err, ErrNotFound) {
		log.Println("error when checking for email collisions:", err)
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	} else if (err == nil || errors.Is(err, ErrNotFound)) && len(matchEmail) > 0 {
		ReturnError(w, http.StatusUnauthorized, "taken", "Username or email already used")
		return
	}
	u, err := NewUser(req.Username, req.Password, req.Email)
	if err != nil {
		log.Println("error creating new user:", err)
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	err = b.userTable.Insert(r.Context(), u)
	if err != nil {
		log.Println("error inserting new user:", err)
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	var ret createUserReturn
	ret.Username = u.Username
	ret.Token, err = b.GenerateJWT(u.toReqUser())
	if err != nil {
		log.Println("error generating token:", err)
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ret)
}

func (b *Backend) deleteUser(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.VerifyHeader(w, r, "management", true)
	if hdr == nil {
		if err != nil {
			log.Println("request key parsing error:", err)
		}
		return
	}
	userID := r.PathValue("userID")
	if userID == "" {
		ReturnError(w, http.StatusBadRequest, "badRequest", "Bad Request")
		return
	}
	err = b.userTable.Remove(r.Context(), userID)
	if err != nil && err != ErrNotFound {
		log.Println("error deleting user:", err)
	}
}

type loginRequest struct {
	Username string
	Password string
}

type loginReturn struct {
	Token    string `json:"token"`
	Error    string `json:"error"`
	ErrorMsg string `json:"errorMsg"`
	Timeout  int64  `json:"timeout"`
}

func (b *Backend) login(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.VerifyHeader(w, r, "user", false)
	if hdr == nil {
		if err != nil {
			log.Println("request key parsing error:", err)
		}
		return
	}
	defer r.Body.Close()
	var req loginRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Username == "" || req.Password == "" {
		ReturnError(w, http.StatusBadRequest, "invalidBody", "Bad request")
		return
	}
	var ret loginReturn
	u, err := b.TryLogin(r.Context(), req.Username, req.Password)
	if err == nil {
		ret.Token, err = b.GenerateJWT(u.toReqUser())
		if err != nil {
			ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
			return
		}
	} else {
		if err == ErrLoginTimeout {
			ret.Error = "timeout"
			ret.ErrorMsg = fmt.Sprint("Timed out for", u.Timeout, "seconds")
			ret.Timeout = u.Timeout
		} else {
			ret.Error = "incorrect"
			ret.ErrorMsg = "Incorrect username or password"
		}
	}
	json.NewEncoder(w).Encode(ret)
}
