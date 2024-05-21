package darkstorm

import "net/http"

type Log struct {
	ID       string
	Platform string
	Date     int
}

func (l Log) GetID() string {
	return l.ID
}

func (b *Backend) log(w http.ResponseWriter, r *http.Request) {
	//TODO
}
