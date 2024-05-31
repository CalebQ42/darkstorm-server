package darkstorm

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

type Backend struct {
	userTable       Table[User]
	keyTable        Table[ApiKey]
	m               *http.ServeMux
	apps            map[string]App
	managementKeyID string
	jwtPriv         ed25519.PrivateKey
	jwtPub          ed25519.PublicKey
	userMutex       sync.RWMutex
}

func NewBackend(keyTable Table[ApiKey], apps ...App) (*Backend, error) {
	b := &Backend{
		keyTable:  keyTable,
		m:         &http.ServeMux{},
		apps:      make(map[string]App),
		userMutex: sync.RWMutex{},
	}
	var hasLog, hasCrash bool
	for i := range apps {
		_, has := b.apps[apps[i].AppID()]
		if has {
			return nil, errors.New("duplicate AppIDs found")
		}
		b.apps[apps[i].AppID()] = apps[i]
		if ext, is := apps[i].(ExtendedApp); is {
			b.m.HandleFunc("/"+apps[i].AppID()+"/", ext.Extension)
		}
		if !hasLog && apps[i].LogTable() != nil {
			hasLog = true
		}
		if !hasCrash && apps[i].CrashTable() != nil {
			hasCrash = true
		}
	}
	if hasLog {
		b.m.HandleFunc("POST /log/{uuid}", b.log)
	}
	if hasCrash {
		b.m.HandleFunc("POST /crash", b.reportCrash)
		b.m.HandleFunc("DELETE /crash/{crashID}", b.deleteCrash)
		b.m.HandleFunc("POST /crash/archive", b.archiveCrash)
	}
	go b.cleanupLoop()
	return b, nil
}

func (b *Backend) cleanupLoop() {
	for range time.Tick(24 * time.Hour) {
		oldTim := time.Now().Add(-30 * 24 * time.Hour)
		old := (oldTim.Year() * 10000) + (int(oldTim.Month()) * 100) + oldTim.Day()
		for _, a := range b.apps {
			tab := a.LogTable()
			if tab == nil {
				continue
			}
			tab.RemoveOldLogs(old)
		}
	}
}

func (b *Backend) EnableManagementKey(managementID string) {
	b.managementKeyID = managementID
	b.m.HandleFunc("DELETE /{appID}/crash/{crashID}", b.managementDeleteCrash)
	b.m.HandleFunc("POST /{appID}/crash/archive", b.managementArchiveCrash)
}

func (b *Backend) AddUserAuth(userTable Table[User], privKey, pubKey []byte) {
	b.userTable = userTable
	b.jwtPriv = privKey
	b.jwtPub = pubKey
}

func (b *Backend) HandleFunc(pattern string, h http.HandlerFunc) {
	b.m.HandleFunc(pattern, h)
}

func (b *Backend) GetApp(a *ApiKey) App {
	return b.apps[a.AppID]
}

type retError struct {
	ErrorCode string `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}

func ReturnError(w http.ResponseWriter, status int, code, msg string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(retError{
		ErrorCode: code,
		ErrorMsg:  msg,
	})
}
