package darkstorm

import "net/http"

type ArchivedCrash struct {
	Error    string
	Stack    string
	Platform string
}

type IndividualCrash struct {
	Platform string
	Error    string
	Stack    string
	Count    int
}

type CrashReport struct {
	ID         string
	Error      string
	FirstLine  string
	Individual []IndividualCrash
}

func (c CrashReport) GetID() string {
	return c.ID
}

type crashReq struct {
	ID       string
	Platform string
	Error    string
	Stack    string
}

func (b *Backend) reportCrash(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.ParseHeader(r)
	if hdr.k == nil || hdr.k.Perm["crash"] {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err != nil {
		//TODO
		return
	}
	//TODO
}

func (b *Backend) deleteCrash(w http.ResponseWriter, r *http.Request) {
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
