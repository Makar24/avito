package models

import (
	"time"
)

// TeamMember представляет участника команды
type TeamMember struct {
	UserID   string `json:"user_id" db:"user_id"`
	Username string `json:"username" db:"username"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

// Team представляет команду
type Team struct {
	TeamName string       `json:"team_name" db:"team_name"`
	Members  []TeamMember `json:"members"`
}

// User представляет пользователя
type User struct {
	UserID   string `json:"user_id" db:"user_id"`
	Username string `json:"username" db:"username"`
	TeamName string `json:"team_name" db:"team_name"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

// PullRequest представляет Pull Request
type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id" db:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name" db:"pull_request_name"`
	AuthorID          string     `json:"author_id" db:"author_id"`
	Status            string     `json:"status" db:"status"` // OPEN или MERGED
	AssignedReviewers []string   `json:"assigned_reviewers"` // Список user_id ревьюверов (0..2)
	CreatedAt         *time.Time `json:"createdAt,omitempty" db:"created_at"`
	MergedAt          *time.Time `json:"mergedAt,omitempty" db:"merged_at"`
}

// PullRequestShort представляет краткую информацию о PR
type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

// CreateTeamRequest запрос на создание команды
type CreateTeamRequest struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

// SetUserActiveRequest запрос на установку активности пользователя
type SetUserActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

// CreatePRRequest запрос на создание PR
type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

// MergePRRequest запрос на merge PR
type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

// ReassignReviewerRequest запрос на переназначение ревьювера
type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

// ErrorDetail детали ошибки
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse ответ с ошибкой
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// TeamResponse ответ с командой
type TeamResponse struct {
	Team Team `json:"team"`
}

// UserResponse ответ с пользователем
type UserResponse struct {
	User User `json:"user"`
}

// PRResponse ответ с PR
type PRResponse struct {
	PR PullRequest `json:"pr"`
}

// ReassignResponse ответ на переназначение
type ReassignResponse struct {
	PR         PullRequest `json:"pr"`
	ReplacedBy string      `json:"replaced_by"`
}

// GetReviewResponse ответ на получение PR пользователя
type GetReviewResponse struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}
