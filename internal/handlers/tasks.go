package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"taskbox/internal/models"
	"time"
)

func (h *Handler) Tasks(w http.ResponseWriter, r *http.Request) {
	user := h.getCurrentUser(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case "GET":
		h.getTasks(w, r, user)
	case "POST":
		h.createTask(w, r, user)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) TaskDetail(w http.ResponseWriter, r *http.Request) {
	user := h.getCurrentUser(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// extract task id from path
	idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		h.getTask(w, r, user, id)
	case "PATCH", "POST": // support POST for html forms
		h.updateTask(w, r, user, id)
	case "DELETE":
		h.deleteTask(w, r, user, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) getTasks(w http.ResponseWriter, r *http.Request, user *models.User) {
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
	}

	data := map[string]interface{}{
		"TasksByPosition": tasksByPosition,
	}

	h.templates.ExecuteTemplate(w, "tasks.html", data)
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request, user *models.User) {
	title := r.FormValue("title")
	position := r.FormValue("position")

	if title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}

	if position == "" {
		position = "inbox"
	}

	// get max matrix_order for this position
	var maxOrder int
	h.db.QueryRow(`
		SELECT COALESCE(MAX(matrix_order), -1) 
		FROM tasks 
		WHERE user_id = ? AND position = ?
	`, user.ID, position).Scan(&maxOrder)

	result, err := h.db.Exec(`
		INSERT INTO tasks (user_id, title, position, matrix_order, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, user.ID, title, position, maxOrder+1)

	if err != nil {
		http.Error(w, "failed to create task", http.StatusInternalServerError)
		return
	}

	taskID, _ := result.LastInsertId()

	// return task card html fragment
	task := models.Task{
		ID:          int(taskID),
		UserID:      user.ID,
		Title:       title,
		Position:    position,
		MatrixOrder: maxOrder + 1,
	}

	data := map[string]interface{}{
		"Task":         task,
		"CommentCount": 0,
	}

	h.templates.ExecuteTemplate(w, "task-card", data)
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request, user *models.User, id int) {
	log.Printf("getTask called for task id %d by user %d", id, user.ID)
	
	var task models.Task
	var tagsJSON sql.NullString
	var dueDateStr sql.NullString
	var description sql.NullString

	err := h.db.QueryRow(`
		SELECT id, title, description, due_date, tags, position, matrix_order, created_at, updated_at
		FROM tasks
		WHERE id = ? AND user_id = ?
	`, id, user.ID).Scan(
		&task.ID,
		&task.Title,
		&description,
		&dueDateStr,
		&tagsJSON,
		&task.Position,
		&task.MatrixOrder,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		log.Printf("task %d not found", id)
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("database error fetching task %d: %v", id, err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
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

	log.Printf("executing template task-detail for task %d: %s", task.ID, task.Title)
	err = h.templates.ExecuteTemplate(w, "task-detail", task)
	if err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("template executed successfully")
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request, user *models.User, id int) {
	// verify task belongs to user
	var userID int
	err := h.db.QueryRow("SELECT user_id FROM tasks WHERE id = ?", id).Scan(&userID)
	if err == sql.ErrNoRows || userID != user.ID {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	// build update query dynamically based on provided fields
	updates := []string{}
	args := []interface{}{}

	if title := r.FormValue("title"); title != "" {
		updates = append(updates, "title = ?")
		args = append(args, title)
	}

	if r.Form.Has("description") {
		description := r.FormValue("description")
		updates = append(updates, "description = ?")
		args = append(args, description)
	}

	if dueDate := r.FormValue("due_date"); dueDate != "" {
		updates = append(updates, "due_date = ?")
		args = append(args, dueDate)
	}

	if tags := r.FormValue("tags"); tags != "" {
		// convert comma-separated to json array
		tagList := strings.Split(tags, ",")
		for i := range tagList {
			tagList[i] = strings.TrimSpace(tagList[i])
		}
		tagsJSON, _ := json.Marshal(tagList)
		updates = append(updates, "tags = ?")
		args = append(args, string(tagsJSON))
	}

	if position := r.FormValue("position"); position != "" {
		updates = append(updates, "position = ?")
		args = append(args, position)
	}

	if order := r.FormValue("matrix_order"); order != "" {
		updates = append(updates, "matrix_order = ?")
		args = append(args, order)
	}

	if len(updates) == 0 {
		http.Error(w, "no fields to update", http.StatusBadRequest)
		return
	}

	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := "UPDATE tasks SET " + strings.Join(updates, ", ") + " WHERE id = ?"
	_, err = h.db.Exec(query, args...)
	if err != nil {
		http.Error(w, "failed to update task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request, user *models.User, id int) {
	result, err := h.db.Exec("DELETE FROM tasks WHERE id = ? AND user_id = ?", id, user.ID)
	if err != nil {
		http.Error(w, "failed to delete task", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
