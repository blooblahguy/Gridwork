package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"net/http"
	"taskbox/internal/models"

	"golang.org/x/crypto/bcrypt"
)

const sessionCookie = "taskbox_session"

// hash password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// check password against hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generate random session token
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// create user
func CreateUser(db *sql.DB, username, password string) (*models.User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	result, err := db.Exec(
		"INSERT INTO users (username, password_hash) VALUES (?, ?)",
		username, hash,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &models.User{
		ID:           int(id),
		Username:     username,
		PasswordHash: hash,
	}, nil
}

// authenticate user
func AuthenticateUser(db *sql.DB, username, password string) (*models.User, error) {
	var user models.User
	err := db.QueryRow(
		"SELECT id, username, password_hash FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash)

	if err != nil {
		return nil, err
	}

	if !CheckPassword(password, user.PasswordHash) {
		return nil, sql.ErrNoRows
	}

	return &user, nil
}

// create session
func CreateSession(db *sql.DB, userID int) (string, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", err
	}

	_, err = db.Exec(
		"INSERT INTO sessions (user_id, token) VALUES (?, ?)",
		userID, token,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

// get user from session token
func GetUserFromSession(db *sql.DB, token string) (*models.User, error) {
	var user models.User
	err := db.QueryRow(`
		SELECT u.id, u.username, u.password_hash 
		FROM users u
		JOIN sessions s ON s.user_id = u.id
		WHERE s.token = ?
	`, token).Scan(&user.ID, &user.Username, &user.PasswordHash)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// delete session
func DeleteSession(db *sql.DB, token string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

// set session cookie
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   0, // session cookie (indefinite)
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// get session token from cookie
func GetSessionToken(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// clear session cookie
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
