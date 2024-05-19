package darkstorm

import (
	"crypto/ed25519"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrApiKeyUnauthorized = errors.New("api key invalid")
	ErrTokenUnauthorized  = errors.New("token invalid")
)

type Backend struct {
	userTable Table[User]
	keyTable  Table[Key]
	m         *http.ServeMux
	jwtPriv   ed25519.PrivateKey
	jwtPub    ed25519.PublicKey
	apps      []App
}

func NewBackend(keyTable Table[Key], apps ...App) (*Backend, error) {
	b := &Backend{
		keyTable: keyTable,
		m:        &http.ServeMux{},
		apps:     apps,
	}
	//TODO: register paths to the mux
	b.startCleanupLoop()
	return b, nil
}

func (b *Backend) AddUserAuth(userTable Table[User], privKey, pubKey []byte) {
	b.userTable = userTable
	b.jwtPriv = privKey
	b.jwtPub = pubKey
}

func (b *Backend) HandleFunc(pattern string, h http.HandlerFunc) {
	b.m.HandleFunc(pattern, h)
}

func (b *Backend) startCleanupLoop() {
	go func() {
		for range time.Tick(6 * time.Hour) {
			//TODO
		}
	}()
}

type ParsedHeader struct {
	u *ReqUser
	k *Key
}

func (b *Backend) ParseHeader(r *http.Request) (ParsedHeader, error) {
	out := ParsedHeader{}
	key := r.Header.Get("X-API-Key")
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if key != "" {
		apiKey, err := b.keyTable.Get(key)
		if err != nil {
			return out, errors.Join(ErrApiKeyUnauthorized, err)
		}
		out.k = &apiKey
	}
	if token != "" && b.userTable != nil {
		t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return b.jwtPub, nil
		}, jwt.WithIssuer("darkstorm.tech"), jwt.WithExpirationRequired())
		if err != nil {
			return out, errors.Join(ErrTokenUnauthorized, err)
		}
		sub, err := t.Claims.GetSubject()
		if err != nil {
			return out, errors.Join(ErrTokenUnauthorized, err)
		}
		usr, err := b.userTable.Get(sub)
		if err != nil{
			return out, errors.Join(ErrTokenUnauthorized, err)
		}
		
	}
	return out, nil
}
