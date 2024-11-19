package blog

import (
	"bytes"
	"log"
	"net/http"
)

func (b *Backend) editorReq(w http.ResponseWriter, r *http.Request) {
	blogForm := new(bytes.Buffer)
	err := b.tmpl.ExecuteTemplate(blogForm, "blogForm", blogFormStruct{
		Blog: Blog{},
	})
	if err != nil {
		log.Println("error using blogForm:", err)
		b.wrapper(w, r, "error", "<p>Server error</p>")
		return
	}
	blogs, err := b.FullBlogList(r.Context())
	if err != nil {
		log.Println("error getting blog list:", err)
		b.wrapper(w, r, "error", "<p>Server error</p>")
		return
	}
	blogPage := new(bytes.Buffer)
	err = b.tmpl.ExecuteTemplate(blogPage, "blogPage", blogPageStruct{
		Selected: "",
		Editor:   blogForm.String(),
		Blogs:    blogs,
	})
	if err != nil {
		log.Println("error using blogPage:", err)
		b.wrapper(w, r, "error", "<p>Server error</p>")
		return
	}
	out := new(bytes.Buffer)
	err = b.tmpl.ExecuteTemplate(out, "editor", editorStruct{
		SelectedPage: "",
		Page:         blogPage.String(),
	})
	if err != nil {
		log.Println("error using editor:", err)
		b.wrapper(w, r, "error", "<p>Server error</p>")
		return
	}
	b.wrapper(w, r, "Editor", out.String())
}
