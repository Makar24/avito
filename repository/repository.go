package repository

import (
	"avito/models"
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// User methods
func (r *Repository) CreateOrUpdateUser(userID, username string, isActive bool) error {
	_, err := r.db.Exec(`
		INSERT INTO users (user_id, username, is_active) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) 
		DO UPDATE SET username = $2, is_active = $3
	`, userID, username, isActive)
	return err
}

func (r *Repository) GetUser(userID string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(`
		SELECT u.user_id, u.username, u.is_active
		FROM users u
		WHERE u.user_id = $1
	`, userID).Scan(&user.UserID, &user.Username, &user.IsActive)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	// Получаем team_name пользователя (берем первую команду)
	err = r.db.QueryRow(`
		SELECT tm.team_name
		FROM team_members tm
		WHERE tm.user_id = $1
		LIMIT 1
	`, userID).Scan(&user.TeamName)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return user, nil
}

func (r *Repository) UpdateUserActivity(userID string, isActive bool) error {
	_, err := r.db.Exec("UPDATE users SET is_active = $1 WHERE user_id = $2", isActive, userID)
	return err
}

// Team methods
func (r *Repository) CreateTeam(teamName string) error {
	_, err := r.db.Exec("INSERT INTO teams (team_name) VALUES ($1)", teamName)
	return err
}

func (r *Repository) TeamExists(teamName string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", teamName).Scan(&exists)
	return exists, err
}

func (r *Repository) GetTeam(teamName string) (*models.Team, error) {
	// Проверяем существование команды
	exists, err := r.TeamExists(teamName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("team not found")
	}

	team := &models.Team{
		TeamName: teamName,
		Members:  []models.TeamMember{},
	}

	// Получаем участников команды
	rows, err := r.db.Query(`
		SELECT u.user_id, u.username, u.is_active
		FROM users u
		INNER JOIN team_members tm ON u.user_id = tm.user_id
		WHERE tm.team_name = $1
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var member models.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, err
		}
		team.Members = append(team.Members, member)
	}

	return team, rows.Err()
}

func (r *Repository) AddUserToTeam(teamName, userID string) error {
	_, err := r.db.Exec("INSERT INTO team_members (team_name, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		teamName, userID)
	return err
}

func (r *Repository) GetUserTeamName(userID string) (string, error) {
	var teamName string
	err := r.db.QueryRow(`
		SELECT team_name
		FROM team_members
		WHERE user_id = $1
		LIMIT 1
	`, userID).Scan(&teamName)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("user is not a member of any team")
	}
	return teamName, err
}

func (r *Repository) GetActiveTeamMembersExcept(teamName, excludeUserID string) ([]*models.User, error) {
	rows, err := r.db.Query(`
		SELECT u.user_id, u.username, u.is_active
		FROM users u
		INNER JOIN team_members tm ON u.user_id = tm.user_id
		WHERE tm.team_name = $1 AND u.is_active = true AND u.user_id != $2
	`, teamName, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.UserID, &user.Username, &user.IsActive); err != nil {
			return nil, err
		}
		user.TeamName = teamName
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *Repository) GetActiveTeamMembers(teamName string) ([]*models.User, error) {
	rows, err := r.db.Query(`
		SELECT u.user_id, u.username, u.is_active
		FROM users u
		INNER JOIN team_members tm ON u.user_id = tm.user_id
		WHERE tm.team_name = $1 AND u.is_active = true
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.UserID, &user.Username, &user.IsActive); err != nil {
			return nil, err
		}
		user.TeamName = teamName
		users = append(users, user)
	}
	return users, rows.Err()
}

// PR methods
func (r *Repository) PRExists(pullRequestID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", pullRequestID).Scan(&exists)
	return exists, err
}

func (r *Repository) CreatePR(pr *models.PullRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var createdAt time.Time
	if pr.CreatedAt != nil {
		createdAt = *pr.CreatedAt
	} else {
		createdAt = time.Now()
	}

	_, err = tx.Exec(`
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, createdAt)
	if err != nil {
		return err
	}

	for _, reviewerID := range pr.AssignedReviewers {
		_, err = tx.Exec("INSERT INTO pr_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)",
			pr.PullRequestID, reviewerID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) GetPR(pullRequestID string) (*models.PullRequest, error) {
	pr := &models.PullRequest{}
	var createdAt sql.NullTime
	var mergedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`, pullRequestID).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &createdAt, &mergedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("PR not found")
	}
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		pr.CreatedAt = &createdAt.Time
	}
	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	// Получаем ревьюверов
	rows, err := r.db.Query("SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id = $1", pullRequestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	return pr, rows.Err()
}

func (r *Repository) UpdatePRStatus(pullRequestID, status string, mergedAt *time.Time) error {
	// Валидация статуса на уровне репозитория для дополнительной защиты
	if status != "OPEN" && status != "MERGED" {
		return fmt.Errorf("invalid status: %s (must be OPEN or MERGED)", status)
	}

	_, err := r.db.Exec(`
		UPDATE pull_requests 
		SET status = $1, merged_at = $2 
		WHERE pull_request_id = $3
	`, status, mergedAt, pullRequestID)
	return err
}

func (r *Repository) ReplaceReviewer(pullRequestID, oldReviewerID, newReviewerID string) error {
	_, err := r.db.Exec(`
		UPDATE pr_reviewers 
		SET reviewer_id = $1 
		WHERE pull_request_id = $2 AND reviewer_id = $3
	`, newReviewerID, pullRequestID, oldReviewerID)
	return err
}

func (r *Repository) GetPRsByReviewer(reviewerID string) ([]*models.PullRequestShort, error) {
	rows, err := r.db.Query(`
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.reviewer_id = $1
		ORDER BY pr.created_at DESC
	`, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []*models.PullRequestShort
	for rows.Next() {
		pr := &models.PullRequestShort{}
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	return prs, rows.Err()
}

// SelectRandomReviewers выбирает случайных ревьюверов из списка (до 2)
func SelectRandomReviewers(candidates []*models.User, count int) []*models.User {
	if count <= 0 || len(candidates) == 0 {
		return []*models.User{}
	}

	if count > len(candidates) {
		count = len(candidates)
	}

	// Создаем копию для перемешивания
	shuffled := make([]*models.User, len(candidates))
	copy(shuffled, candidates)

	// Перемешиваем
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}
