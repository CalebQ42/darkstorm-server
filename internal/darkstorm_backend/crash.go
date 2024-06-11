package darkstorm

import (
	"encoding/json"
	"log"
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
	hdr, err := b.VerifyHeader(w, r, "crash", false)
	if hdr == nil {
		if err == nil {
			log.Println("request key parsing error:", err)
		}
		return
	}
	ap := b.GetApp(hdr.Key)
	defer r.Body.Close()
	var crash IndividualCrash
	err = json.NewDecoder(r.Body).Decode(&crash)
	if err != nil || crash.Platform == "" || crash.Error == "" || crash.Stack == "" {
		ReturnError(w, http.StatusBadRequest, "invalidBody", "Bad request")
		return
	}
	tab := ap.CrashTable()
	if tab == nil {
		log.Printf("key %v has crash permission, but app does not have a crash table", hdr.Key.AppID)
		ReturnError(w, http.StatusInternalServerError, "misconfigured", "Server misconfigured")
		return
	}
	if !tab.IsArchived(crash) {
		err = tab.InsertCrash(crash)
		if err != nil {
			log.Println("crash insertion error:", err)
			ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
			return
		}
	}
	w.WriteHeader(http.StatusCreated)
}

func (b *Backend) deleteCrash(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.VerifyHeader(w, r, "management", false)
	if hdr == nil {
		if err == nil {
			log.Println("request key parsing error:", err)
		}
		return
	}
	crashID := r.PathValue("crashID")
	if crashID == "" {
		ReturnError(w, http.StatusBadRequest, "badRequest", "Bad request")
		return
	}
	b.actualCrashDelete(w, b.GetApp(hdr.Key), crashID)
}

func (b *Backend) managementDeleteCrash(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.VerifyHeader(w, r, "management", true)
	if hdr == nil {
		if err == nil {
			log.Println("request key parsing error:", err)
		}
		return
	}
	crashID := r.PathValue("crashID")
	if crashID == "" {
		ReturnError(w, http.StatusBadRequest, "badRequest", "Bad request")
		return
	}
	appID := r.PathValue("appID")
	ap := b.apps[appID]
	if ap == nil || appID == "" {
		ReturnError(w, http.StatusBadRequest, "badRequest", "Bad request")
		return
	}
	b.actualCrashDelete(w, ap, crashID)
}

func (b *Backend) actualCrashDelete(w http.ResponseWriter, ap App, crashID string) {
	crash := ap.CrashTable()
	if crash == nil {
		ReturnError(w, http.StatusInternalServerError, "misconfigured", "Server Misconfigured")
		return
	}
	err := crash.Remove(crashID)
	if err != nil && err != ErrNotFound {
		log.Println("error when deleting crash:", err)
	}
}

func (b *Backend) archiveCrash(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.VerifyHeader(w, r, "management", false)
	if hdr == nil {
		if err == nil {
			log.Println("request key parsing error:", err)
		}
		return
	}
	defer r.Body.Close()
	var toArchive ArchivedCrash
	err = json.NewDecoder(r.Body).Decode(&toArchive)
	if err != nil || toArchive.Platform == "" || toArchive.Error == "" || toArchive.Stack == "" {
		ReturnError(w, http.StatusBadRequest, "invalidBody", "Bad request")
		return
	}
	b.actualCrashArchive(w, b.GetApp(hdr.Key), toArchive)
}

func (b *Backend) managementArchiveCrash(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.VerifyHeader(w, r, "management", true)
	if hdr == nil {
		if err == nil {
			log.Println("request key parsing error:", err)
		}
		return
	}
	appID := r.PathValue("appID")
	ap := b.apps[appID]
	if ap == nil || appID == "" {
		ReturnError(w, http.StatusBadRequest, "badRequest", "Bad request")
		return
	}
	defer r.Body.Close()
	var toArchive ArchivedCrash
	err = json.NewDecoder(r.Body).Decode(&toArchive)
	if err != nil || toArchive.Platform == "" || toArchive.Error == "" || toArchive.Stack == "" {
		ReturnError(w, http.StatusBadRequest, "invalidBody", "Bad request")
		return
	}
	b.actualCrashArchive(w, ap, toArchive)
}

func (b *Backend) actualCrashArchive(w http.ResponseWriter, ap App, toArchive ArchivedCrash) {
	crash := ap.CrashTable()
	if crash == nil {
		ReturnError(w, http.StatusInternalServerError, "misconfigured", "Server Misconfigured")
		return
	}
	err := crash.Archive(toArchive)
	if err != nil {
		log.Println()
	}
}
