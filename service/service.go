package service

import (
	"avito/models"
	"avito/repository"
	"fmt"
	"time"
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

// CreateTeam создает команду с участниками (создает/обновляет пользователей)
func (s *Service) CreateTeam(req *models.CreateTeamRequest) (*models.Team, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.TeamName == "" {
		return nil, fmt.Errorf("team name cannot be empty")
	}

	// Проверяем, существует ли команда
	exists, err := s.repo.TeamExists(req.TeamName)
	if err != nil {
		return nil, fmt.Errorf("failed to check team existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("TEAM_EXISTS: team_name already exists")
	}

	// Создаем команду
	if err := s.repo.CreateTeam(req.TeamName); err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	// Создаем/обновляем пользователей и добавляем их в команду
	for _, member := range req.Members {
		if err := s.repo.CreateOrUpdateUser(member.UserID, member.Username, member.IsActive); err != nil {
			return nil, fmt.Errorf("failed to create/update user %s: %w", member.UserID, err)
		}

		if err := s.repo.AddUserToTeam(req.TeamName, member.UserID); err != nil {
			return nil, fmt.Errorf("failed to add user to team: %w", err)
		}
	}

	// Возвращаем созданную команду
	return s.repo.GetTeam(req.TeamName)
}

// GetTeam получает команду по имени
func (s *Service) GetTeam(teamName string) (*models.Team, error) {
	if teamName == "" {
		return nil, fmt.Errorf("team name cannot be empty")
	}

	team, err := s.repo.GetTeam(teamName)
	if err != nil {
		return nil, fmt.Errorf("NOT_FOUND: %w", err)
	}

	return team, nil
}

// SetUserActive устанавливает флаг активности пользователя
func (s *Service) SetUserActive(userID string, isActive bool) (*models.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	// Проверяем существование пользователя
	_, err := s.repo.GetUser(userID)
	if err != nil {
		return nil, fmt.Errorf("NOT_FOUND: user not found")
	}

	// Обновляем активность
	if err := s.repo.UpdateUserActivity(userID, isActive); err != nil {
		return nil, fmt.Errorf("failed to update user activity: %w", err)
	}

	// Возвращаем обновленного пользователя
	return s.repo.GetUser(userID)
}

// CreatePR создает новый PR и автоматически назначает ревьюверов
func (s *Service) CreatePR(req *models.CreatePRRequest) (*models.PullRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.PullRequestID == "" {
		return nil, fmt.Errorf("pull request ID cannot be empty")
	}
	if req.PullRequestName == "" {
		return nil, fmt.Errorf("pull request name cannot be empty")
	}
	if req.AuthorID == "" {
		return nil, fmt.Errorf("author ID cannot be empty")
	}

	// Проверяем, существует ли PR
	exists, err := s.repo.PRExists(req.PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to check PR existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("PR_EXISTS: PR id already exists")
	}

	// Проверяем существование автора
	_, err = s.repo.GetUser(req.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("NOT_FOUND: author not found")
	}

	// Получаем команду автора
	teamName, err := s.repo.GetUserTeamName(req.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("NOT_FOUND: author is not a member of any team")
	}

	// Получаем активных участников команды, исключая автора
	candidates, err := s.repo.GetActiveTeamMembersExcept(teamName, req.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	// Выбираем до 2 случайных ревьюверов
	reviewers := repository.SelectRandomReviewers(candidates, 2)
	reviewerIDs := make([]string, len(reviewers))
	for i, r := range reviewers {
		reviewerIDs[i] = r.UserID
	}

	now := time.Now()
	pr := &models.PullRequest{
		PullRequestID:     req.PullRequestID,
		PullRequestName:   req.PullRequestName,
		AuthorID:          req.AuthorID,
		Status:            "OPEN",
		AssignedReviewers: reviewerIDs,
		CreatedAt:         &now,
	}

	if err := s.repo.CreatePR(pr); err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return s.repo.GetPR(req.PullRequestID)
}

// ReassignReviewer переназначает ревьювера
func (s *Service) ReassignReviewer(pullRequestID, oldUserID string) (*models.PullRequest, string, error) {
	if pullRequestID == "" {
		return nil, "", fmt.Errorf("pull request ID cannot be empty")
	}
	if oldUserID == "" {
		return nil, "", fmt.Errorf("old user ID cannot be empty")
	}

	// Получаем PR
	pr, err := s.repo.GetPR(pullRequestID)
	if err != nil {
		return nil, "", fmt.Errorf("NOT_FOUND: PR not found")
	}

	// Проверяем, что PR не в статусе MERGED
	if pr.Status == "MERGED" {
		return nil, "", fmt.Errorf("PR_MERGED: cannot reassign on merged PR")
	}

	// Проверяем, что старый ревьювер действительно назначен
	found := false
	for _, rid := range pr.AssignedReviewers {
		if rid == oldUserID {
			found = true
			break
		}
	}
	if !found {
		return nil, "", fmt.Errorf("NOT_ASSIGNED: reviewer is not assigned to this PR")
	}

	// Получаем команду старого ревьювера
	teamName, err := s.repo.GetUserTeamName(oldUserID)
	if err != nil {
		return nil, "", fmt.Errorf("NOT_FOUND: old reviewer is not a member of any team")
	}

	// Получаем активных участников команды
	candidates, err := s.repo.GetActiveTeamMembers(teamName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get team members: %w", err)
	}

	// Исключаем уже назначенных ревьюверов и автора
	availableCandidates := []*models.User{}
	for _, candidate := range candidates {
		isAssigned := false
		for _, rid := range pr.AssignedReviewers {
			if candidate.UserID == rid {
				isAssigned = true
				break
			}
		}
		if !isAssigned && candidate.UserID != pr.AuthorID && candidate.UserID != oldUserID {
			availableCandidates = append(availableCandidates, candidate)
		}
	}

	if len(availableCandidates) == 0 {
		return nil, "", fmt.Errorf("NO_CANDIDATE: no active replacement candidate in team")
	}

	// Выбираем случайного нового ревьювера
	newReviewers := repository.SelectRandomReviewers(availableCandidates, 1)
	if len(newReviewers) == 0 {
		return nil, "", fmt.Errorf("NO_CANDIDATE: failed to select new reviewer")
	}

	newReviewerID := newReviewers[0].UserID

	// Заменяем ревьювера
	if err := s.repo.ReplaceReviewer(pullRequestID, oldUserID, newReviewerID); err != nil {
		return nil, "", fmt.Errorf("failed to replace reviewer: %w", err)
	}

	// Возвращаем обновленный PR
	updatedPR, err := s.repo.GetPR(pullRequestID)
	if err != nil {
		return nil, "", err
	}

	return updatedPR, newReviewerID, nil
}

// MergePR выполняет merge PR (идемпотентная операция)
func (s *Service) MergePR(pullRequestID string) (*models.PullRequest, error) {
	if pullRequestID == "" {
		return nil, fmt.Errorf("pull request ID cannot be empty")
	}

	pr, err := s.repo.GetPR(pullRequestID)
	if err != nil {
		return nil, fmt.Errorf("NOT_FOUND: PR not found")
	}

	// Если уже merged, просто возвращаем текущее состояние (идемпотентность)
	if pr.Status == "MERGED" {
		return pr, nil
	}

	// Выполняем merge
	// Валидация статуса уже выполнена выше (проверка на "MERGED")
	now := time.Now()
	if err := s.repo.UpdatePRStatus(pullRequestID, "MERGED", &now); err != nil {
		return nil, fmt.Errorf("failed to merge PR: %w", err)
	}

	return s.repo.GetPR(pullRequestID)
}

// GetReview возвращает список PR, назначенных ревьюверу
func (s *Service) GetReview(userID string) (*models.GetReviewResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	// Проверяем существование пользователя
	_, err := s.repo.GetUser(userID)
	if err != nil {
		return nil, fmt.Errorf("NOT_FOUND: user not found")
	}

	prs, err := s.repo.GetPRsByReviewer(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs: %w", err)
	}

	// Преобразуем []*models.PullRequestShort в []models.PullRequestShort
	pullRequests := make([]models.PullRequestShort, len(prs))
	for i, pr := range prs {
		pullRequests[i] = *pr
	}

	return &models.GetReviewResponse{
		UserID:       userID,
		PullRequests: pullRequests,
	}, nil
}
