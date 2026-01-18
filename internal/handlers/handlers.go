package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"taskbox/internal/auth"
	"taskbox/internal/models"
)

type Handler struct {
	db        *sql.DB
	templates *template.Template
	devMode   bool
}

func New(db *sql.DB, devMode bool) *Handler {
	log.Println("loading templates...")
	
	// parse all templates recursively
	tmpl := template.New("").Funcs(template.FuncMap{
		"dict": dict,
	})
	
	// collect all template files
	patterns := []string{
		"templates/pages/*.html",
		"templates/parts/tasks/*.html",
		"templates/parts/comments/*.html",
	}
	
	allFiles := []string{}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			log.Printf("error globbing %s: %v", pattern, err)
			continue
		}
		log.Printf("found %d templates matching %s", len(matches), pattern)
		allFiles = append(allFiles, matches...)
	}
	
	// parse all files together
	if len(allFiles) > 0 {
		var err error
		tmpl, err = tmpl.ParseFiles(allFiles...)
		if err != nil {
			log.Fatalf("error parsing templates: %v", err)
		}
	}
	
	log.Printf("loaded %d templates", len(allFiles))
	log.Println("available templates:", tmpl.DefinedTemplates())

	return &Handler{
		db:        db,
		templates: tmpl,
		devMode:   devMode,
	}
}

// helper to create a map for template data
func dict(values ...interface{}) map[string]interface{} {
	if len(values)%2 != 0 {
		panic("dict requires even number of arguments")
	}
	m := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			panic("dict keys must be strings")
		}
		m[key] = values[i+1]
	}
	return m
}

// get current user from session
func (h *Handler) getCurrentUser(r *http.Request) *models.User {
	token := auth.GetSessionToken(r)
	if token == "" {
		return nil
	}

	user, err := auth.GetUserFromSession(h.db, token)
	if err != nil {
		return nil
	}

	return user
}

// require authentication middleware
func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := h.getCurrentUser(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// dev mode helper
func (h *Handler) DevMode() bool {
	return h.devMode
}
