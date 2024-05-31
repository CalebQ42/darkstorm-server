package darkstorm

import (
	"encoding/json"
	"errors"
	"net/http"
)

type ArchivedCrash struct {
	Error    string `json:"error" bson:"error"`
	Stack    string `json:"stack" bson:"stack"`
	Platform string `json:"platform" bson:"platform"`
}

type IndividualCrash struct {
	Platform string `json:"platform" bson:"platform"`
	Error    string `json:"error" bson:"error"`
	Stack    string `json:"stack" bson:"stack"`
	Count    int    `json:"count" bson:"count"`
}

type CrashReport struct {
	ID         string            `json:"id" bson:"_id"`
	Error      string            `json:"error" bson:"error"`
	FirstLine  string            `json:"firstLine" bson:"firstLine"`
	Individual []IndividualCrash `json:"individual" bson:"individual"`
}

func (c CrashReport) GetID() string {
	return c.ID
}

func (b *Backend) reportCrash(w http.ResponseWriter, r *http.Request) {
	var ap App
	hdr, err := b.ParseHeader(r)
	if hdr.k != nil {
		ap = b.GetApp(hdr.k)
	}
	if ap == nil || hdr.k.Perm["crash"] || errors.Is(err, ErrApiKeyUnauthorized) {
		ReturnError(w, http.StatusUnauthorized, "invalidKey", "Application not authorized")
		return
	} else if err != nil {
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	defer r.Body.Close()
	var crash IndividualCrash
	err = json.NewDecoder(r.Body).Decode(&crash)
	if err != nil || crash.Platform == "" || crash.Error == "" || crash.Stack == "" {
		ReturnError(w, http.StatusBadRequest, "invalidBody", "Bad request")
		return
	}
	tab := ap.CrashTable()
	if tab == nil {
		ReturnError(w, http.StatusInternalServerError, "misconfigured", "Server misconfigured")
		return
	}
	if !tab.IsArchived(crash) {
		err = tab.InsertCrash(crash)
		if err != nil {
			ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
			return
		}
	}
	w.WriteHeader(http.StatusCreated)
}

func (b *Backend) deleteCrash(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.ParseHeader(r)
	if hdr.k == nil || hdr.k.Perm["management"] || errors.Is(err, ErrApiKeyUnauthorized) {
		ReturnError(w, http.StatusUnauthorized, "invalidKey", "Application not authorized")
		return
	} else if err != nil {
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	//TODO
}

func (b *Backend) managementDeleteCrash(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.ParseHeader(r)
	if hdr.k == nil || hdr.k.Perm["management"] || errors.Is(err, ErrApiKeyUnauthorized) {
		ReturnError(w, http.StatusUnauthorized, "invalidKey", "Application not authorized")
		return
	} else if err != nil {
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	//TODO
}

func (b *Backend) actualCrashDelete(w http.ResponseWriter, ap App, crashID string) {}

func (b *Backend) archiveCrash(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.ParseHeader(r)
	if hdr.k == nil || hdr.k.Perm["management"] {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err != nil {
		//TODO
		return
	}
	//TODO
}

func (b *Backend) managementArchiveCrash(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.ParseHeader(r)
	if hdr.k == nil || hdr.k.Perm["management"] {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err != nil {
		//TODO
		return
	}
	//TODO
}

func (b *Backend) actualCrashArchive(w http.ResponseWriter, ap App, toArchive ArchivedCrash) {}
