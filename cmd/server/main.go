package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"taskbox/internal/database"
	"taskbox/internal/handlers"
	"taskbox/internal/scss"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// check if in dev mode
	devMode := os.Getenv("DEV_MODE") == "true"

	// initialize database
	db, err := sql.Open("sqlite3", "./taskbox.db")
	if err != nil {
		log.Fatal("failed to open database:", err)
	}
	defer db.Close()

	// run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("failed to run migrations:", err)
	}

	// setup handlers
	mux := http.NewServeMux()
	handlers := handlers.New(db, devMode)

	// static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// health check for dev mode reload
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// routes
	mux.HandleFunc("/", handlers.Index)
	mux.HandleFunc("/register", handlers.Register)
	mux.HandleFunc("/login", handlers.Login)
	mux.HandleFunc("/logout", handlers.Logout)
	mux.HandleFunc("/tasks", handlers.Tasks)
	mux.HandleFunc("/tasks/", handlers.TaskDetail)
	mux.HandleFunc("/comments/", handlers.Comments)

	// start scss watcher in background
	go scss.Watch("./scss", "./static/css")
	log.Println("scss watcher started")

	if devMode {
		log.Println("running in dev mode - browser auto-reload enabled")
	}

	log.Println("server starting on :1234")
	if err := http.ListenAndServe(":1234", mux); err != nil {
		log.Fatal("server failed:", err)
	}
}
