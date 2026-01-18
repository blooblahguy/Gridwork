package handlers

import (
	"net/http"
	"taskbox/internal/auth"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.templates.ExecuteTemplate(w, "register.html", map[string]interface{}{
			"DevMode": h.devMode,
		})
		return
	}

	// handle POST
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		h.templates.ExecuteTemplate(w, "register.html", map[string]interface{}{
			"Error":   "username and password required",
			"DevMode": h.devMode,
		})
		return
	}

	user, err := auth.CreateUser(h.db, username, password)
	if err != nil {
		h.templates.ExecuteTemplate(w, "register.html", map[string]interface{}{
			"Error":   "username already exists",
			"DevMode": h.devMode,
		})
		return
	}

	// create session
	token, err := auth.CreateSession(h.db, user.ID)
	if err != nil {
		h.templates.ExecuteTemplate(w, "register.html", map[string]interface{}{
			"Error":   "failed to create session",
			"DevMode": h.devMode,
		})
		return
	}

	auth.SetSessionCookie(w, token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"DevMode": h.devMode,
		})
		return
	}

	// handle POST
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := auth.AuthenticateUser(h.db, username, password)
	if err != nil {
		h.templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"Error":   "invalid credentials",
			"DevMode": h.devMode,
		})
		return
	}

	// create session
	token, err := auth.CreateSession(h.db, user.ID)
	if err != nil {
		h.templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"Error":   "failed to create session",
			"DevMode": h.devMode,
		})
		return
	}

	auth.SetSessionCookie(w, token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token := auth.GetSessionToken(r)
	if token != "" {
		auth.DeleteSession(h.db, token)
	}
	auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
