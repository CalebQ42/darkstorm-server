package main

import (
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
)

//go:embed embed
var editorFS embed.FS

const loginPage = `
<form id="loginForm" hx-post="/login">
	<label for="username">Username:</label>
	<input name="username" id="usernameInput" onkeydown="return event.key != 'Enter';"></input>
	<label for="password">Password:</label>
	<input name="password" type="password" id="passwordInput"></input>
	<div id="formResult"></div>
	<button id="loginButton" type="submit">Login</button>
</form>
`

func LoginPage(w http.ResponseWriter, r *http.Request) {
	sendContent(w, r, loginPage, "", "")
}

func TrueLogin(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		sendContent(w, r, "<p>Bad request</p>", "", "")
	}
	u, err := back.TryLogin(r.Context(), r.URL.Query().Get("username"), r.URL.Query().Get("password"))
	if err != nil {
		if err == backend.ErrLoginTimeout {
			sendContent(w, r, fmt.Sprint("<p>Timed out for", time.Unix(u.Timeout, 0).Sub(time.Now()), "</p>"), "", "")
		} else if err == backend.ErrLoginTimeout {
			sendContent(w, r, "<p>Username or password invalid</p>", "", "")
		} else {
			log.Println("error trying to login:", err)
			sendContent(w, r, "<p>Server error</p>", "", "")
		}
		return
	}
	tok, err := back.GenerateJWT(u.ToReqUser())
	if err != nil {
		log.Println("error trying to generate JWT:", err)
		sendContent(w, r, "<p>Server error</p>", "", "")
		return
	}
	w.Header().Set("Set-Cookie", "blogAuthToken="+tok+"; Secure; Max-Age=43170") // Max-Age is 11.5 hours. JWTs are valid for 12 hours.
	sendContent(w, r, "<p hx-get='/editor' hx-push-url='true' hx-trigger='load' hx-target='#content'>Successful Login</p>", "", "")
}

func Editor(w http.ResponseWriter, r *http.Request) {
	authCookie, err := r.Cookie("blogAuthToken")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		if err != http.ErrNoCookie {
			log.Println("error getting auth cookie:", err)
		}
		editorRedirect(w, r, "/login")
		return
	}
	usr, err := back.VerifyUser(r.Context(), authCookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		if err != backend.ErrTokenUnauthorized {
			log.Println("error authorizing JWT token:", err)
		}
		editorRedirect(w, r, "/login")
		return
	}
	page, err := editorFS.Open("embed/editor.html")
	defer page.Close()
	if err != nil {
		log.Println("error getting editor.html:", err)
		sendContent(w, r, "error getting page", "", "")
		return
	}
	_, err = io.ReadAll(page)
	if err != nil {
		log.Println("error reading editor.html:", err)
		sendContent(w, r, "error getting page", "", "")
		return
	}
	sendContent(w, r, "<p>Hello there, "+usr.Username+"</p>", "", "")
}

func editorRedirect(w http.ResponseWriter, r *http.Request, path string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Location", `{"path": "`+path+`", "target":"#content"}`)
		return
	}
	http.Redirect(w, r, "https://darkstorm.tech"+path, http.StatusSeeOther)
}
