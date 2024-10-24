package backend

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrApiKeyUnauthorized = errors.New("api key present but invalid")
	ErrTokenUnauthorized  = errors.New("token present but invalid")
)

type ApiKey struct {
	Perm  map[string]bool `json:"perm" bson:"perm"`
	ID    string          `json:"id" bson:"_id" valkey:",key"`
	AppID string          `json:"appID" bson:"appID"`
	Death int64           `json:"death" bson:"death"`
}

func (k ApiKey) GetID() string {
	return k.ID
}

type ParsedHeader struct {
	User *ReqestUser
	Key  *ApiKey
}

// Parses the X-API-Key and Authorization headers. If the API Key provided but invalid (either due to expiring or isn't found), ErrApiKeyUnauthorized is returned.
// If the Authorization header is present but invalid, ErrTokenUnauthorized is returned.
// NOTE: An invalid apiKey will cause a nil return, but a invalid token will not. Token parsing is only
func (b *Backend) ParseHeader(r *http.Request) (*ParsedHeader, error) {
	out := &ParsedHeader{}
	key := r.Header.Get("X-API-Key")

	//TODO: Remove legacy code
	if key == "" {
		key = r.URL.Query().Get("key")
	}

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if key != "" {
		apiKey, err := b.keyTable.Get(r.Context(), key)
		if err == ErrNotFound {
			return nil, ErrApiKeyUnauthorized
		} else if err != nil {
			return nil, err
		}
		if apiKey.Death > 0 && time.Unix(apiKey.Death, 0).Before(time.Now()) {
			return nil, ErrApiKeyUnauthorized
		}
		out.Key = apiKey
	}
	if token != "" && b.userTable != nil {
		t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return b.jwtPub, nil
		}, jwt.WithIssuer("darkstorm.tech"), jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{"EdDSA"}))
		if err != nil {
			return out, errors.Join(ErrTokenUnauthorized, err)
		}
		exp, _ := t.Claims.GetExpirationTime()
		if exp.Time.Before(time.Now()) {
			return out, ErrTokenUnauthorized
		}
		sub, err := t.Claims.GetSubject()
		if err == jwt.ErrInvalidKey {
			return out, ErrTokenUnauthorized
		} else if err != nil {
			return out, errors.Join(ErrTokenUnauthorized, err)
		}
		usr, err := b.userTable.Get(r.Context(), sub)
		if err == jwt.ErrInvalidKey {
			return out, ErrTokenUnauthorized
		} else if err != nil {
			return out, errors.Join(ErrTokenUnauthorized, err)
		}
		iss, err := t.Claims.GetIssuedAt()
		if err == jwt.ErrInvalidKey {
			return out, ErrTokenUnauthorized
		} else if err != nil {
			return out, errors.Join(ErrTokenUnauthorized, err)
		}
		if usr.PasswordChange > 0 && iss.Time.Before(time.Unix(usr.PasswordChange, 0)) {
			return out, ErrTokenUnauthorized
		}
		out.User = usr.toReqUser()
	}
	return out, nil
}

// Similiar to ParseHeader, but with key checking and automatic error returns. Guarentess Backend.GetApp is non-nil
// Checks that the key is a management key (not management permission and if allowManagement is true) or that it has the necessary permission.
// If the check if failed, ReturnError will be called and the returned *ParsedHeader will be nil.
// If token is present but invalid, no error will be returned just ParsedHeader.User will be nil.
// The error return will only be populated on "internal" errors and should *probably* be logged.
//
// This function does not check the Key's appID so after calling VerifyHeader it's recommended to check the Key's appID.
func (b *Backend) VerifyHeader(w http.ResponseWriter, r *http.Request, keyPerm string, allowManagementKey bool) (*ParsedHeader, error) {
	hdr, err := b.ParseHeader(r)
	if hdr == nil || hdr.Key == nil {
		if err == ErrApiKeyUnauthorized {
			fmt.Println("yo1")
			ReturnError(w, http.StatusUnauthorized, "invalidKey", "Application not authorized")
			return nil, nil
		}
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return nil, err
	}
	if err != nil && !errors.Is(err, ErrTokenUnauthorized) {
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return nil, err
	}
	if hdr.Key.AppID == b.managementKeyID {
		if allowManagementKey {
			return hdr, nil
		} else {
			fmt.Println("yo2")
			ReturnError(w, http.StatusUnauthorized, "invalidKey", "Application not authorized")
			return nil, nil
		}
	}
	if _, ok := b.apps[hdr.Key.AppID]; !ok {
		fmt.Println("yo3")
		ReturnError(w, http.StatusUnauthorized, "invalidKey", "Application not authorized")
		return nil, errors.New("server misconfigured, appID present in DB, but App not added to backend")
	}
	return hdr, nil
}
