package darkstorm

import (
	"log"
	"net/http"
)

type CountLog struct {
	ID       string
	Platform string
	Date     int
}

func (c CountLog) GetID() string {
	return c.ID
}

func (b *Backend) countLog(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.VerifyHeader(w, r, "count", false)
	if hdr == nil {
		if err == nil {
			log.Println("request key parsing error:", err)
		}
		return
	}
	//TODO
}

func (b *Backend) getCount(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.VerifyHeader(w, r, "management", true)
	if hdr == nil {
		if err == nil {
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
	//TODO
}
