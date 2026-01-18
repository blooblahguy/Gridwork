package handlers

import (
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) Comments(w http.ResponseWriter, r *http.Request) {
	user := h.getCurrentUser(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// extract task id from path: /comments/{taskID}
	path := strings.TrimPrefix(r.URL.Path, "/comments/")
	taskID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		h.GetComments(w, r, user, taskID)
	case "POST":
		h.AddComment(w, r, user, taskID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
