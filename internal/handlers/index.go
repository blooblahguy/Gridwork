package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"taskbox/internal/models"
	"time"
)

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	log.Println("index handler called")
	
	user := h.getCurrentUser(r)
	if user == nil {
		log.Println("no user, redirecting to login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	log.Printf("user authenticated: %s", user.Username)

	// fetch all tasks for user
	rows, err := h.db.Query(`
		SELECT 
			t.id, t.title, t.description, t.due_date, t.tags, t.position, 
			t.matrix_order, t.created_at, t.updated_at,
			(SELECT COUNT(*) FROM comments c WHERE c.task_id = t.id) as comment_count
		FROM tasks t
		WHERE t.user_id = ?
		ORDER BY t.position, t.matrix_order
	`, user.ID)
	if err != nil {
		log.Printf("database error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// organize tasks by position
	tasksByPosition := map[string][]map[string]interface{}{
		"inbox":    {},
		"do":       {},
		"decide":   {},
		"delegate": {},
		"delete":   {},
		"archive":  {},
	}

	taskCount := 0
	for rows.Next() {
		var task models.Task
		var tagsJSON sql.NullString
		var dueDateStr sql.NullString
		var description sql.NullString
		var commentCount int

		err := rows.Scan(
			&task.ID,
			&task.Title,
			&description,
			&dueDateStr,
			&tagsJSON,
			&task.Position,
			&task.MatrixOrder,
			&task.CreatedAt,
			&task.UpdatedAt,
			&commentCount,
		)
		if err != nil {
			continue
		}

		task.UserID = user.ID
		if description.Valid {
			task.Description = description.String
		}
		if dueDateStr.Valid {
			t, _ := time.Parse(time.RFC3339, dueDateStr.String)
			task.DueDate = &t
		}
		if tagsJSON.Valid && tagsJSON.String != "" {
			json.Unmarshal([]byte(tagsJSON.String), &task.Tags)
		}

		taskData := map[string]interface{}{
			"Task":         task,
			"CommentCount": commentCount,
		}

		tasksByPosition[task.Position] = append(tasksByPosition[task.Position], taskData)
		taskCount++
	}

	log.Printf("loaded %d tasks", taskCount)

	data := map[string]interface{}{
		"User":            user,
		"TasksByPosition": tasksByPosition,
		"DevMode":         h.devMode,
	}

	log.Println("executing template: index.html")
	err = h.templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Println("template executed successfully")
}
