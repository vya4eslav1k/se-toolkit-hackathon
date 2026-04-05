package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/maksimslavik/inno-se-toolkit-pet/core/internal/model"
)

type PollRepository interface {
	CreatePoll(ctx context.Context, poll *model.Poll) error
	CreateQuestion(ctx context.Context, question *model.Question) error
	CreateOption(ctx context.Context, option *model.Option) error
	GetPollByID(ctx context.Context, id string) (*model.Poll, error)
	GetQuestionsByPollID(ctx context.Context, pollID string) ([]model.Question, error)
	GetOptionsByQuestionID(ctx context.Context, questionID string) ([]model.Option, error)
	HasVoted(ctx context.Context, pollID, voterID, questionID string) (bool, error)
	CreateVote(ctx context.Context, vote *model.Vote, questionID, optionID string) error
	GetVoteCount(ctx context.Context, pollID string) (int, error)
	GetOptionVoteCount(ctx context.Context, optionID string) (int, error)
	GetPublicPolls(ctx context.Context, limit int) ([]model.Poll, error)
}

type PollService struct {
	repo PollRepository
}

func NewPollService(repo PollRepository) *PollService {
	return &PollService{repo: repo}
}

func (s *PollService) CreatePoll(req *model.CreatePollRequest, createdBy string) (*model.Poll, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if len(req.Questions) == 0 {
		return nil, fmt.Errorf("at least one question is required")
	}

	poll := &model.Poll{
		ID:               uuid.New().String(),
		Title:            req.Title,
		CreatedBy:        createdBy,
		Visibility:       req.Visibility,
		ResultVisibility: req.ResultVisibility,
	}

	if poll.Visibility == "" {
		poll.Visibility = model.VisibilityPublic
	}
	if poll.ResultVisibility == "" {
		poll.ResultVisibility = model.ResultVisibilityAlways
	}

	ctx := context.Background()
	if err := s.repo.CreatePoll(ctx, poll); err != nil {
		return nil, err
	}

	for _, qReq := range req.Questions {
		if qReq.Text == "" {
			return nil, fmt.Errorf("question text is required")
		}
		if len(qReq.Options) < 2 {
			return nil, fmt.Errorf("at least 2 options required for question: %s", qReq.Text)
		}

		question := &model.Question{
			ID:     uuid.New().String(),
			PollID: poll.ID,
			Text:   qReq.Text,
		}

		if err := s.repo.CreateQuestion(ctx, question); err != nil {
			return nil, err
		}

		for _, optText := range qReq.Options {
			if optText == "" {
				return nil, fmt.Errorf("option text cannot be empty")
			}

			option := &model.Option{
				ID:         uuid.New().String(),
				QuestionID: question.ID,
				Text:       optText,
			}

			if err := s.repo.CreateOption(ctx, option); err != nil {
				return nil, err
			}
		}

		poll.Questions = append(poll.Questions, *question)
	}

	return poll, nil
}

func (s *PollService) GetPoll(id string) (*model.Poll, error) {
	ctx := context.Background()
	poll, err := s.repo.GetPollByID(ctx, id)
	if err != nil {
		return nil, err
	}

	questions, err := s.repo.GetQuestionsByPollID(ctx, id)
	if err != nil {
		return nil, err
	}

	for i := range questions {
		options, err := s.repo.GetOptionsByQuestionID(ctx, questions[i].ID)
		if err != nil {
			return nil, err
		}
		questions[i].Options = options
	}

	poll.Questions = questions
	return poll, nil
}

func (s *PollService) SubmitVote(pollID, voterID string, answers []model.AnswerRequest) error {
	ctx := context.Background()
	poll, err := s.repo.GetPollByID(ctx, pollID)
	if err != nil {
		return err
	}

	for _, answer := range answers {
		hasVoted, err := s.repo.HasVoted(ctx, pollID, voterID, answer.QuestionID)
		if err != nil {
			return err
		}
		if hasVoted {
			return fmt.Errorf("already voted for question %s", answer.QuestionID)
		}

		vote := &model.Vote{
			ID:      uuid.New().String(),
			PollID:  pollID,
			VoterID: voterID,
		}

		if err := s.repo.CreateVote(ctx, vote, answer.QuestionID, answer.OptionID); err != nil {
			return err
		}
	}

	_ = poll
	return nil
}

func (s *PollService) GetPollResults(pollID, viewerID string) (*model.PollResult, error) {
	ctx := context.Background()
	poll, err := s.repo.GetPollByID(ctx, pollID)
	if err != nil {
		return nil, err
	}

	questions, err := s.repo.GetQuestionsByPollID(ctx, pollID)
	if err != nil {
		return nil, err
	}

	totalVotes, err := s.repo.GetVoteCount(ctx, pollID)
	if err != nil {
		return nil, err
	}

	canView := s.canViewResults(poll, viewerID, totalVotes)

	for i := range questions {
		options, err := s.repo.GetOptionsByQuestionID(ctx, questions[i].ID)
		if err != nil {
			return nil, err
		}

		for j := range options {
			count, err := s.repo.GetOptionVoteCount(ctx, options[j].ID)
			if err != nil {
				return nil, err
			}
			options[j].Votes = count
		}

		questions[i].Options = options
	}

	poll.Questions = questions

	return &model.PollResult{
		Poll:       *poll,
		TotalVotes: totalVotes,
		CanView:    canView,
	}, nil
}

func (s *PollService) GetPublicPolls(limit int, viewerID string) ([]model.Poll, error) {
	ctx := context.Background()
	polls, err := s.repo.GetPublicPolls(ctx, limit)
	if err != nil {
		return nil, err
	}

	// Filter out unlisted polls for non-authors
	filtered := make([]model.Poll, 0, len(polls))
	for _, poll := range polls {
		if poll.Visibility == model.VisibilityUnlisted && poll.CreatedBy != viewerID {
			continue
		}
		filtered = append(filtered, poll)
	}

	return filtered, nil
}

func (s *PollService) HasVotedForQuestion(pollID, voterID, questionID string) (bool, error) {
	ctx := context.Background()
	return s.repo.HasVoted(ctx, pollID, voterID, questionID)
}

func (s *PollService) canViewResults(poll *model.Poll, viewerID string, totalVotes int) bool {
	switch poll.ResultVisibility {
	case model.ResultVisibilityAlways:
		return true
	case model.ResultVisibilityAfterVote:
		return totalVotes > 0
	case model.ResultVisibilityAuthorOnly:
		return viewerID == poll.CreatedBy
	default:
		return true
	}
}
