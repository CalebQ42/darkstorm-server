package blog

import "net/http"


func (b *Backend) editorPage(w http.ResponseWriter, r *http.Request) {
	pag := r.PathValue("page")
	if
}
