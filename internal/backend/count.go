package backend

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type CountLog struct {
	ID       string `json:"id" bson:"_id"`
	Platform string `json:"platform" bson:"platform"`
	Date     int    `json:"date" bson:"date"`
}

func (c CountLog) GetID() string {
	return c.ID
}

type countLogReq struct {
	ID       string
	Platform string
}

func (b *Backend) countLog(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.VerifyHeader(w, r, "count", false)
	if hdr == nil {
		if err != nil {
			log.Println("request key parsing error:", err)
		}
		return
	}
	defer r.Body.Close()
	var req countLogReq
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Platform == "" {
		if r.URL.Query().Get("platform") != "" {
			//TODO: remove legacy code
			req.Platform = r.URL.Query().Get("platform")
			req.ID = r.URL.Query().Get("id")
		} else {
			ReturnError(w, http.StatusBadRequest, "invalidBody", "Bad request")
			return
		}
	}
	ap := b.GetApp(hdr.Key)
	count := ap.CountTable()
	if count == nil {
		ReturnError(w, http.StatusInternalServerError, "misconfigured", "Server Misconfigured")
		return
	}
	curDate := getDate(time.Now())
	if req.ID == "" {
		err = addToCountTable(r.Context(), w, count, req.Platform, curDate)
		if err != nil {
			log.Println("error adding to count table:", err)
		}
		return
	}
	l, err := count.Get(r.Context(), req.ID)
	if err == ErrNotFound {
		err = addToCountTable(r.Context(), w, count, req.Platform, curDate)
		if err != nil {
			log.Println("error adding to count table:", err)
		}
		return
	}
	if l.Date >= curDate {
		json.NewEncoder(w).Encode(map[string]string{"id": req.ID})
		w.WriteHeader(http.StatusCreated)
		return
	}
	err = count.PartUpdate(r.Context(), req.ID, map[string]any{"date": curDate})
	if err != nil {
		log.Println("error updating count log:", err)
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": req.ID})
}

func addToCountTable(ctx context.Context, w http.ResponseWriter, c CountTable, platform string, curDate int) error {
	id, err := uuid.NewV7()
	if err != nil {
		log.Println("error generating new log UUID:", err)
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return err
	}
	err = c.Insert(ctx, CountLog{
		ID:       id.String(),
		Platform: platform,
		Date:     curDate,
	})
	if err != nil {
		log.Println("error inserting new count log:", err)
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return err
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id.String()})
	return nil
}

func (b *Backend) getCount(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.VerifyHeader(w, r, "management", true)
	if hdr == nil {
		if err != nil {
			log.Println("request key parsing error:", err)
		}
		return
	}
	var ap App
	if hdr.Key.AppID == b.managementKeyID {
		ap = b.apps[r.PathValue("appID")]
		if ap == nil {
			ReturnError(w, http.StatusBadRequest, "badRequest", "Bad request")
			return
		}
	} else {
		ap = b.GetApp(hdr.Key)
	}
	count := ap.CountTable()
	if count == nil {
		ReturnError(w, http.StatusBadRequest, "badRequest", "Trying to get user count on app that doesn't have a count table")
		return
	}
	out, err := count.Count(r.Context(), r.URL.Query().Get("platform"))
	if err != nil {
		log.Println("error getting count:", err)
		ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	json.NewEncoder(w).Encode(map[string]int{"count": out})
}
