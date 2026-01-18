package models

import "time"

type User struct {
	ID           int
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type Session struct {
	ID        int
	UserID    int
	Token     string
	CreatedAt time.Time
}

type Task struct {
	ID          int
	UserID      int
	Title       string
	Description string
	DueDate     *time.Time
	Tags        []string
	Position    string
	MatrixOrder int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Comment struct {
	ID        int
	TaskID    int
	UserID    int
	Content   string
	CreatedAt time.Time
}
