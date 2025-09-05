package backend

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	ErrAPIKeyUnauthorized = errors.New("api key present but invalid")
	ErrTokenUnauthorized  = errors.New("token present but invalid")
)

type APIKey struct {
	Perm           map[string]bool `json:"perm" bson:"perm"`
	ID             string          `json:"id" bson:"_id" valkey:",key"`
	AppID          string          `json:"appID" bson:"appID"`
	Death          int64           `json:"death" bson:"death"`
	AllowedOrigins []string        `json:"allowedOrigins" bson:"allowedOrigins"`
}

func (k APIKey) GetID() string {
	return k.ID
}

type ParsedHeader struct {
	User *ReqestUser
	Key  *APIKey
}

// Parses the X-API-Key and Authorization headers. If the API Key provided but invalid (either due to expiring or isn't found), ErrApiKeyUnauthorized is returned.
// If the Authorization header is present but invalid, ErrTokenUnauthorized is returned.
// NOTE: An invalid apiKey will cause a nil return, but a invalid token will not. Token parsing is only
func (b *Backend) ParseHeader(r *http.Request) (*ParsedHeader, error) {
	out := &ParsedHeader{}
	key := r.Header.Get("X-API-Key")

	if key != "" {
		apiKey, err := b.keyTable.Get(r.Context(), key)
		if err == ErrNotFound {
			return nil, ErrAPIKeyUnauthorized
		} else if err != nil {
			return nil, err
		}
		if apiKey.Death > 0 && time.Unix(apiKey.Death, 0).Before(time.Now()) {
			return nil, ErrAPIKeyUnauthorized
		}
		out.Key = &apiKey
	} else {
		fmt.Println("origin:", r.Header.Get("origin"))
		keys, err := b.keyTable.Find(r.Context(), map[string]any{"allowedOrigins": r.Header.Get("origin")})
		if err == ErrNotFound {
			return nil, ErrAPIKeyUnauthorized
		} else if err != nil {
			return nil, err
		}
		if keys[0].Death > 0 && time.Unix(keys[0].Death, 0).Before(time.Now()) {
			return nil, ErrAPIKeyUnauthorized
		}
		out.Key = &keys[0]
	}
	if b.userTable == nil || r.Header.Get("Authorization") == "" {
		return out, nil
	}
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		return out, nil
	}
	usr, err := b.VerifyUser(r.Context(), token)
	if err != nil {
		return out, err
	}
	out.User = usr.ToReqUser()
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
		if err == ErrAPIKeyUnauthorized {
			ReturnError(w, http.StatusUnauthorized, "invalidKey", "Application not authorized")
			return nil, nil
		}
		ReturnError(w, http.StatusUnauthorized, "noKey", "No API Key provided")
		return nil, err
	}
	if err != nil && !errors.Is(err, ErrTokenUnauthorized) {
		log.Println("error parsing header:", err)
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return nil, err
	}
	if hdr.Key.AppID == b.managementKeyID {
		if allowManagementKey {
			return hdr, nil
		} else {
			ReturnError(w, http.StatusUnauthorized, "invalidKey", "Application not authorized")
			return nil, nil
		}
	}
	if _, ok := b.apps[hdr.Key.AppID]; !ok {
		ReturnError(w, http.StatusUnauthorized, "invalidKey", "Application not authorized")
		return nil, errors.New("server misconfigured, appID present in DB, but App not added to backend")
	}
	return hdr, nil
}
