package web

import (
	"net/http"

	"github.com/bennycio/bundle/api"
	"github.com/bennycio/bundle/internal/gate"
)

func signupHandlerFunc(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		r.ParseForm()

		user := &api.User{
			Username: r.FormValue("username"),
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}

		gs := gate.NewGateService("", "")
		err := gs.InsertUser(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		token, err := newAuthToken(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tokenCookie := newAccessCookie(token)
		http.SetCookie(w, tokenCookie)

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

		return
	}

	user, err := getUserFromCookie(r)

	td := TemplateData{}

	if err == nil {
		td.User = user
	}

	err = tpl.ExecuteTemplate(w, "register", td)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
