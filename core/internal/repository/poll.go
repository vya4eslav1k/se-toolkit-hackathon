package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/maksimslavik/inno-se-toolkit-pet/core/internal/model"
)

type PollRepository struct {
	db *pgx.Conn
}

func NewPollRepository(db *pgx.Conn) *PollRepository {
	return &PollRepository{db: db}
}

func (r *PollRepository) CreatePoll(ctx context.Context, poll *model.Poll) error {
	query := `INSERT INTO polls (id, title, created_by, visibility, result_visibility) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, poll.ID, poll.Title, poll.CreatedBy, poll.Visibility, poll.ResultVisibility)
	if err != nil {
		return fmt.Errorf("failed to create poll: %w", err)
	}
	return nil
}

func (r *PollRepository) CreateQuestion(ctx context.Context, question *model.Question) error {
	query := `INSERT INTO questions (id, poll_id, text) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, question.ID, question.PollID, question.Text)
	if err != nil {
		return fmt.Errorf("failed to create question: %w", err)
	}
	return nil
}

func (r *PollRepository) CreateOption(ctx context.Context, option *model.Option) error {
	query := `INSERT INTO options (id, question_id, text) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, option.ID, option.QuestionID, option.Text)
	if err != nil {
		return fmt.Errorf("failed to create option: %w", err)
	}
	return nil
}

func (r *PollRepository) GetPollByID(ctx context.Context, id string) (*model.Poll, error) {
	var poll model.Poll
	query := `SELECT id, title, created_at, created_by, visibility, result_visibility FROM polls WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&poll.ID, &poll.Title, &poll.CreatedAt, &poll.CreatedBy, &poll.Visibility, &poll.ResultVisibility)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("poll not found")
		}
		return nil, fmt.Errorf("failed to get poll: %w", err)
	}
	return &poll, nil
}

func (r *PollRepository) GetQuestionsByPollID(ctx context.Context, pollID string) ([]model.Question, error) {
	query := `SELECT id, poll_id, text FROM questions WHERE poll_id = $1`
	rows, err := r.db.Query(ctx, query, pollID)
	if err != nil {
		return nil, fmt.Errorf("failed to get questions: %w", err)
	}
	defer rows.Close()

	var questions []model.Question
	for rows.Next() {
		var q model.Question
		if err := rows.Scan(&q.ID, &q.PollID, &q.Text); err != nil {
			return nil, fmt.Errorf("failed to scan question: %w", err)
		}
		questions = append(questions, q)
	}
	return questions, nil
}

func (r *PollRepository) GetOptionsByQuestionID(ctx context.Context, questionID string) ([]model.Option, error) {
	query := `SELECT id, question_id, text FROM options WHERE question_id = $1`
	rows, err := r.db.Query(ctx, query, questionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}
	defer rows.Close()

	var options []model.Option
	for rows.Next() {
		var o model.Option
		if err := rows.Scan(&o.ID, &o.QuestionID, &o.Text); err != nil {
			return nil, fmt.Errorf("failed to scan option: %w", err)
		}
		options = append(options, o)
	}
	return options, nil
}

func (r *PollRepository) HasVoted(ctx context.Context, pollID, voterID, questionID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM votes WHERE poll_id = $1 AND voter_id = $2 AND question_id = $3)`
	err := r.db.QueryRow(ctx, query, pollID, voterID, questionID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check vote: %w", err)
	}
	return exists, nil
}

func (r *PollRepository) CreateVote(ctx context.Context, vote *model.Vote, questionID, optionID string) error {
	query := `INSERT INTO votes (id, poll_id, voter_id, question_id, option_id) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, vote.ID, vote.PollID, vote.VoterID, questionID, optionID)
	if err != nil {
		return fmt.Errorf("failed to create vote: %w", err)
	}
	return nil
}

func (r *PollRepository) GetVoteCount(ctx context.Context, pollID string) (int, error) {
	var count int
	query := `SELECT COUNT(DISTINCT voter_id) FROM votes WHERE poll_id = $1`
	err := r.db.QueryRow(ctx, query, pollID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get vote count: %w", err)
	}
	return count, nil
}

func (r *PollRepository) GetOptionVoteCount(ctx context.Context, optionID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM votes WHERE option_id = $1`
	err := r.db.QueryRow(ctx, query, optionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get option vote count: %w", err)
	}
	return count, nil
}

func (r *PollRepository) GetPublicPolls(ctx context.Context, limit int) ([]model.Poll, error) {
	query := `SELECT id, title, created_at, created_by, visibility, result_visibility FROM polls ORDER BY created_at DESC LIMIT $1`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get public polls: %w", err)
	}
	defer rows.Close()

	var polls []model.Poll
	for rows.Next() {
		var p model.Poll
		if err := rows.Scan(&p.ID, &p.Title, &p.CreatedAt, &p.CreatedBy, &p.Visibility, &p.ResultVisibility); err != nil {
			return nil, fmt.Errorf("failed to scan poll: %w", err)
		}
		polls = append(polls, p)
	}
	return polls, nil
}
