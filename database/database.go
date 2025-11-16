package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func NewDB(connStr string) (*DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) InitSchema() error {
	queries := []string{
		// Таблица пользователей
		`CREATE TABLE IF NOT EXISTS users (
			user_id VARCHAR(255) PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT true
		)`,

		// Таблица команд (team_name используется как id)
		`CREATE TABLE IF NOT EXISTS teams (
			team_name VARCHAR(255) PRIMARY KEY
		)`,

		// Таблица участников команд
		`CREATE TABLE IF NOT EXISTS team_members (
			team_name VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			PRIMARY KEY (team_name, user_id),
			FOREIGN KEY (team_name) REFERENCES teams(team_name) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
		)`,

		// Таблица Pull Request'ов
		`CREATE TABLE IF NOT EXISTS pull_requests (
			pull_request_id VARCHAR(255) PRIMARY KEY,
			pull_request_name VARCHAR(255) NOT NULL,
			author_id VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'OPEN',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			merged_at TIMESTAMP,
			FOREIGN KEY (author_id) REFERENCES users(user_id)
		)`,

		// Таблица ревьюверов PR
		`CREATE TABLE IF NOT EXISTS pr_reviewers (
			pull_request_id VARCHAR(255) NOT NULL,
			reviewer_id VARCHAR(255) NOT NULL,
			PRIMARY KEY (pull_request_id, reviewer_id),
			FOREIGN KEY (pull_request_id) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
			FOREIGN KEY (reviewer_id) REFERENCES users(user_id) ON DELETE CASCADE
		)`,

		// Индексы для оптимизации
		`CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_team_members_team ON team_members(team_name)`,
		`CREATE INDEX IF NOT EXISTS idx_team_members_user ON team_members(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pr_author ON pull_requests(author_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pr_status ON pull_requests(status)`,
		`CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pr ON pr_reviewers(pull_request_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}
