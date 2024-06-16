package backend

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"log"
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
		if !hasLog && apps[i].CountTable() != nil {
			hasLog = true
		}
		if !hasCrash && apps[i].CrashTable() != nil {
			hasCrash = true
		}
	}
	if hasLog {
		b.m.HandleFunc("POST /count", b.countLog)
		b.m.HandleFunc("GET /count", b.getCount)
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
		old := getDate(time.Now().Add(-30 * 24 * time.Hour))
		var err error
		for _, a := range b.apps {
			log.Printf("Removing logs for %v", a.AppID())
			tab := a.CountTable()
			if tab == nil {
				continue
			}
			err = tab.RemoveOldLogs(old)
			if err != nil {
				log.Printf("error removing old logs for %v: %v\n", a.AppID(), err)
			}
		}
	}
}

func (b *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.m.ServeHTTP(w, r)
}

func getDate(t time.Time) int {
	return (t.Year() * 10000) + (int(t.Month()) * 100) + t.Day()
}

func (b *Backend) EnableManagementKey(managementID string) {
	b.managementKeyID = managementID
	b.m.HandleFunc("DELETE /{appID}/crash/{crashID}", b.managementDeleteCrash)
	b.m.HandleFunc("POST /{appID}/crash/archive", b.managementArchiveCrash)
	b.m.HandleFunc("GET /{appID}/count", b.getCount)
}

func (b *Backend) AddUserAuth(userTable Table[User], privKey, pubKey []byte) {
	b.userTable = userTable
	b.jwtPriv = privKey
	b.jwtPub = pubKey
	b.m.HandleFunc("POST /user/create", b.createUser)
	b.m.HandleFunc("DELETE /user/{userID}", b.deleteUser)
	b.m.HandleFunc("POST /user/login", b.login)
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
