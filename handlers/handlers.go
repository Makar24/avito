package handlers

import (
	"avito/models"
	"avito/service"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type Handlers struct {
	service *service.Service
}

func NewHandlers(svc *service.Service) *Handlers {
	return &Handlers{service: svc}
}

func (h *Handlers) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) respondError(w http.ResponseWriter, status int, code, message string) {
	h.respondJSON(w, status, models.ErrorResponse{
		Error: models.ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func (h *Handlers) parseError(err error) (int, string, string) {
	errStr := err.Error()

	// Парсим код ошибки из формата "CODE: message"
	parts := strings.SplitN(errStr, ": ", 2)
	if len(parts) == 2 {
		code := parts[0]
		message := parts[1]

		switch code {
		case "TEAM_EXISTS":
			return http.StatusBadRequest, "TEAM_EXISTS", message
		case "PR_EXISTS":
			return http.StatusConflict, "PR_EXISTS", message
		case "PR_MERGED":
			return http.StatusConflict, "PR_MERGED", message
		case "NOT_ASSIGNED":
			return http.StatusConflict, "NOT_ASSIGNED", message
		case "NO_CANDIDATE":
			return http.StatusConflict, "NO_CANDIDATE", message
		case "NOT_FOUND":
			return http.StatusNotFound, "NOT_FOUND", message
		}
	}

	// По умолчанию
	return http.StatusBadRequest, "ERROR", errStr
}

// AddTeam создает команду с участниками
func (h *Handlers) AddTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "ERROR", "Method not allowed")
		return
	}

	var req models.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		h.respondError(w, http.StatusBadRequest, "ERROR", "Invalid request body")
		return
	}

	team, err := h.service.CreateTeam(&req)
	if err != nil {
		log.Printf("Error creating team: %v", err)
		status, code, msg := h.parseError(err)
		h.respondError(w, status, code, msg)
		return
	}

	h.respondJSON(w, http.StatusCreated, models.TeamResponse{Team: *team})
}

// GetTeam получает команду по имени
func (h *Handlers) GetTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "ERROR", "Method not allowed")
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.respondError(w, http.StatusBadRequest, "ERROR", "team_name parameter is required")
		return
	}

	team, err := h.service.GetTeam(teamName)
	if err != nil {
		log.Printf("Error getting team: %v", err)
		status, code, msg := h.parseError(err)
		h.respondError(w, status, code, msg)
		return
	}

	h.respondJSON(w, http.StatusOK, *team)
}

// SetUserActive устанавливает флаг активности пользователя
func (h *Handlers) SetUserActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "ERROR", "Method not allowed")
		return
	}

	var req models.SetUserActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		h.respondError(w, http.StatusBadRequest, "ERROR", "Invalid request body")
		return
	}

	user, err := h.service.SetUserActive(req.UserID, req.IsActive)
	if err != nil {
		log.Printf("Error setting user activity: %v", err)
		status, code, msg := h.parseError(err)
		h.respondError(w, status, code, msg)
		return
	}

	h.respondJSON(w, http.StatusOK, models.UserResponse{User: *user})
}

// CreatePR создает новый PR
func (h *Handlers) CreatePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "ERROR", "Method not allowed")
		return
	}

	var req models.CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		h.respondError(w, http.StatusBadRequest, "ERROR", "Invalid request body")
		return
	}

	pr, err := h.service.CreatePR(&req)
	if err != nil {
		log.Printf("Error creating PR: %v", err)
		status, code, msg := h.parseError(err)
		h.respondError(w, status, code, msg)
		return
	}

	h.respondJSON(w, http.StatusCreated, models.PRResponse{PR: *pr})
}

// ReassignReviewer переназначает ревьювера
func (h *Handlers) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "ERROR", "Method not allowed")
		return
	}

	var req models.ReassignReviewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		h.respondError(w, http.StatusBadRequest, "ERROR", "Invalid request body")
		return
	}

	pr, replacedBy, err := h.service.ReassignReviewer(req.PullRequestID, req.OldUserID)
	if err != nil {
		log.Printf("Error reassigning reviewer: %v", err)
		status, code, msg := h.parseError(err)
		h.respondError(w, status, code, msg)
		return
	}

	h.respondJSON(w, http.StatusOK, models.ReassignResponse{
		PR:         *pr,
		ReplacedBy: replacedBy,
	})
}

// MergePR выполняет merge PR
func (h *Handlers) MergePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "ERROR", "Method not allowed")
		return
	}

	var req models.MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		h.respondError(w, http.StatusBadRequest, "ERROR", "Invalid request body")
		return
	}

	pr, err := h.service.MergePR(req.PullRequestID)
	if err != nil {
		log.Printf("Error merging PR: %v", err)
		status, code, msg := h.parseError(err)
		h.respondError(w, status, code, msg)
		return
	}

	h.respondJSON(w, http.StatusOK, models.PRResponse{PR: *pr})
}

// GetReview возвращает список PR, назначенных ревьюверу
func (h *Handlers) GetReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "ERROR", "Method not allowed")
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.respondError(w, http.StatusBadRequest, "ERROR", "user_id parameter is required")
		return
	}

	resp, err := h.service.GetReview(userID)
	if err != nil {
		log.Printf("Error getting review: %v", err)
		status, code, msg := h.parseError(err)
		h.respondError(w, status, code, msg)
		return
	}

	h.respondJSON(w, http.StatusOK, *resp)
}
