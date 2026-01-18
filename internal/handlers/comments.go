package handlers

import (
	"net/http"
	"strings"
	"taskbox/internal/models"
)

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request, user *models.User, taskID int) {
	// verify task belongs to user
	var userID int
	err := h.db.QueryRow("SELECT user_id FROM tasks WHERE id = ?", taskID).Scan(&userID)
	if err != nil || userID != user.ID {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	// get comments
	rows, err := h.db.Query(`
		SELECT c.id, c.task_id, c.user_id, c.content, c.created_at, u.username
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.task_id = ?
		ORDER BY c.created_at ASC
	`, taskID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	comments := []map[string]interface{}{}
	for rows.Next() {
		var comment models.Comment
		var username string

		err := rows.Scan(
			&comment.ID,
			&comment.TaskID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&username,
		)
		if err != nil {
			continue
		}

		comments = append(comments, map[string]interface{}{
			"Comment":  comment,
			"Username": username,
		})
	}

	data := map[string]interface{}{
		"Comments": comments,
		"TaskID":   taskID,
	}

	h.templates.ExecuteTemplate(w, "comments-list", data)
}

func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request, user *models.User, taskID int) {
	// verify task belongs to user
	var userID int
	err := h.db.QueryRow("SELECT user_id FROM tasks WHERE id = ?", taskID).Scan(&userID)
	if err != nil || userID != user.ID {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" {
		http.Error(w, "content required", http.StatusBadRequest)
		return
	}

	// insert comment
	result, err := h.db.Exec(`
		INSERT INTO comments (task_id, user_id, content)
		VALUES (?, ?, ?)
	`, taskID, user.ID, content)
	if err != nil {
		http.Error(w, "failed to add comment", http.StatusInternalServerError)
		return
	}

	commentID, _ := result.LastInsertId()

	// return comment html
	data := map[string]interface{}{
		"Comment": models.Comment{
			ID:      int(commentID),
			TaskID:  taskID,
			UserID:  user.ID,
			Content: content,
		},
		"Username": user.Username,
	}

	h.templates.ExecuteTemplate(w, "comment-item", data)
}
