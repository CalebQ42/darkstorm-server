package darkstorm

import "net/http"

type Backend struct {
	http.ServeMux
}
