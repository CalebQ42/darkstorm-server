package darkstorm

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	Username string `json:"username"`
	Token    string `json:"token"`
}

func (b *Backend) CreateUser(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.VerifyHeader(w, r, "user", false)
	if hdr == nil {
		if err == nil {
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
	b.userMutex.Lock()
	defer b.userMutex.Unlock()
	matchUsername, err := b.userTable.Find(map[string]any{"username": req.Username})
	if err != nil && !errors.Is(err, ErrNotFound) {
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	} else if (err == nil || errors.Is(err, ErrNotFound)) && len(matchUsername) > 0 {
		ReturnError(w, http.StatusUnauthorized, "taken", "Username or email already used")
		return
	}
	matchEmail, err := b.userTable.Find(map[string]any{"email": req.Email})
	if err != nil && !errors.Is(err, ErrNotFound) {
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	} else if (err == nil || errors.Is(err, ErrNotFound)) && len(matchEmail) > 0 {
		ReturnError(w, http.StatusUnauthorized, "taken", "Username or email already used")
		return
	}
	u, err := NewUser(req.Username, req.Password, req.Email)
	if err != nil {
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	err = b.userTable.Insert(u)
	if err != nil {
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	var ret createUserReturn
	ret.Username = u.Username
	ret.Token, err = b.generateJWT(u.toReqUser())
	if err != nil {
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ret)
}

type loginRequest struct {
	Username string
	Password string
}

type loginReturn struct {
	Token   string `json:"token"`
	Error   string `json:"error"`
	Timeout int64  `json:"timeout"`
}

func (b *Backend) Login(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.ParseHeader(r)
	if hdr.Key == nil || !hdr.Key.Perm["user"] || errors.Is(err, ErrApiKeyUnauthorized) {
		ReturnError(w, http.StatusUnauthorized, "invalidKey", "Application not authorized")
		return
	} else if err != nil {
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	defer r.Body.Close()
	var req loginRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Username == "" || req.Password == "" {
		ReturnError(w, http.StatusBadRequest, "invalidBody", "Bad request")
		return
	}
	b.userMutex.RLock()
	defer b.userMutex.RUnlock()
	var ret loginReturn
	users, err := b.userTable.Find(map[string]any{"username": req.Username})
	if errors.Is(err, ErrNotFound) || len(users) != 1 {
		ret.Error = "invalid"
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ret)
		return
	}
	u := users[0]
	if time.Unix(u.Timeout, 0).After(time.Now()) {
		ret.Error = "timeout"
		ret.Timeout = time.Now().Unix() - u.Timeout
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ret)
		return
	}
	hash, err := u.HashPassword(req.Password)
	if err != nil {
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	if u.Password == hash {
		ret.Token, err = b.generateJWT(u.toReqUser())
		if err != nil {
			ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
			return
		}
		json.NewEncoder(w).Encode(ret)
	} else {
		ret.Error = "invalid"
		upd := map[string]any{"fails": u.Fails + 1}
		if (u.Fails+1)%3 == 0 {
			minutes := 3 ^ ((u.Fails / 3) - 1)
			timeout := time.Now().Add(time.Duration(minutes) * time.Minute).Unix()
			upd["timeout"] = timeout
			ret.Timeout = timeout - time.Now().Unix()
		}
		b.userTable.PartUpdate(u.ID, upd)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ret)
	}
}
