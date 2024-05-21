package darkstorm

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrApiKeyUnauthorized = errors.New("api key present but invalid")
	ErrTokenUnauthorized  = errors.New("token present but invalid")
)

type ParsedHeader struct {
	u *ReqUser
	k *ApiKey
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
		if apiKey.Death > 0 && time.Unix(apiKey.Death, 0).Before(time.Now()) {
			return out, ErrApiKeyUnauthorized
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
		exp, _ := t.Claims.GetExpirationTime()
		if exp.Time.Before(time.Now()) {
			return out, ErrTokenUnauthorized
		}
		sub, err := t.Claims.GetSubject()
		if err != nil {
			return out, errors.Join(ErrTokenUnauthorized, err)
		}
		usr, err := b.userTable.Get(sub)
		if err != nil {
			return out, errors.Join(ErrTokenUnauthorized, err)
		}
		iss, err := t.Claims.GetIssuedAt()
		if err != nil {
			return out, errors.Join(ErrTokenUnauthorized, err)
		}
		if usr.PasswordChange > 0 && iss.Time.Before(time.Unix(usr.PasswordChange, 0)) {
			return out, ErrTokenUnauthorized
		}
		out.u = usr.toReqUser()
	}
	return out, nil
}
