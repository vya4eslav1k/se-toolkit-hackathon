package service

import (
	"context"
	"testing"

	"github.com/maksimslavik/inno-se-toolkit-pet/core/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository
type MockPollRepository struct {
	mock.Mock
}

func (m *MockPollRepository) CreatePoll(ctx context.Context, poll *model.Poll) error {
	args := m.Called(ctx, poll)
	return args.Error(0)
}

func (m *MockPollRepository) CreateQuestion(ctx context.Context, question *model.Question) error {
	args := m.Called(ctx, question)
	return args.Error(0)
}

func (m *MockPollRepository) CreateOption(ctx context.Context, option *model.Option) error {
	args := m.Called(ctx, option)
	return args.Error(0)
}

func (m *MockPollRepository) GetPollByID(ctx context.Context, id string) (*model.Poll, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Poll), args.Error(1)
}

func (m *MockPollRepository) GetQuestionsByPollID(ctx context.Context, pollID string) ([]model.Question, error) {
	args := m.Called(ctx, pollID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Question), args.Error(1)
}

func (m *MockPollRepository) GetOptionsByQuestionID(ctx context.Context, questionID string) ([]model.Option, error) {
	args := m.Called(ctx, questionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Option), args.Error(1)
}

func (m *MockPollRepository) HasVoted(ctx context.Context, pollID, voterID, questionID string) (bool, error) {
	args := m.Called(ctx, pollID, voterID, questionID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPollRepository) CreateVote(ctx context.Context, vote *model.Vote, questionID, optionID string) error {
	args := m.Called(ctx, vote, questionID, optionID)
	return args.Error(0)
}

func (m *MockPollRepository) GetVoteCount(ctx context.Context, pollID string) (int, error) {
	args := m.Called(ctx, pollID)
	return args.Int(0), args.Error(1)
}

func (m *MockPollRepository) GetOptionVoteCount(ctx context.Context, optionID string) (int, error) {
	args := m.Called(ctx, optionID)
	return args.Int(0), args.Error(1)
}

func (m *MockPollRepository) GetPublicPolls(ctx context.Context, limit int) ([]model.Poll, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Poll), args.Error(1)
}

func TestPollService_CreatePoll_Success(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	req := &model.CreatePollRequest{
		Title:            "Test Poll",
		ResultVisibility: model.ResultVisibilityAlways,
		Questions: []model.QuestionRequest{
			{
				Text:    "Question 1?",
				Options: []string{"Option A", "Option B"},
			},
		},
	}

	mockRepo.On("CreatePoll", mock.Anything, mock.AnythingOfType("*model.Poll")).Return(nil)
	mockRepo.On("CreateQuestion", mock.Anything, mock.AnythingOfType("*model.Question")).Return(nil)
	mockRepo.On("CreateOption", mock.Anything, mock.AnythingOfType("*model.Option")).Return(nil)

	poll, err := service.CreatePoll(req, "user123")

	assert.NoError(t, err)
	assert.NotNil(t, poll)
	assert.Equal(t, "Test Poll", poll.Title)
	assert.Equal(t, "user123", poll.CreatedBy)
	assert.Equal(t, model.ResultVisibilityAlways, poll.ResultVisibility)
	assert.Len(t, poll.Questions, 1)

	mockRepo.AssertExpectations(t)
}

func TestPollService_CreatePoll_EmptyTitle(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	req := &model.CreatePollRequest{
		Title:            "",
		ResultVisibility: model.ResultVisibilityAlways,
		Questions: []model.QuestionRequest{
			{
				Text:    "Question 1?",
				Options: []string{"Option A", "Option B"},
			},
		},
	}

	poll, err := service.CreatePoll(req, "user123")

	assert.Error(t, err)
	assert.Nil(t, poll)
	assert.Contains(t, err.Error(), "title is required")
}

func TestPollService_CreatePoll_NoQuestions(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	req := &model.CreatePollRequest{
		Title:            "Test Poll",
		ResultVisibility: model.ResultVisibilityAlways,
		Questions:        []model.QuestionRequest{},
	}

	poll, err := service.CreatePoll(req, "user123")

	assert.Error(t, err)
	assert.Nil(t, poll)
	assert.Contains(t, err.Error(), "at least one question is required")
}

func TestPollService_CreatePoll_LessThanTwoOptions(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	req := &model.CreatePollRequest{
		Title:            "Test Poll",
		ResultVisibility: model.ResultVisibilityAlways,
		Questions: []model.QuestionRequest{
			{
				Text:    "Question 1?",
				Options: []string{"Only one option"},
			},
		},
	}

	mockRepo.On("CreatePoll", mock.Anything, mock.AnythingOfType("*model.Poll")).Return(nil)

	poll, err := service.CreatePoll(req, "user123")

	assert.Error(t, err)
	assert.Nil(t, poll)
	assert.Contains(t, err.Error(), "at least 2 options required")
	mockRepo.AssertExpectations(t)
}

func TestPollService_CreatePoll_EmptyQuestionText(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	req := &model.CreatePollRequest{
		Title:            "Test Poll",
		ResultVisibility: model.ResultVisibilityAlways,
		Questions: []model.QuestionRequest{
			{
				Text:    "",
				Options: []string{"Option A", "Option B"},
			},
		},
	}

	mockRepo.On("CreatePoll", mock.Anything, mock.AnythingOfType("*model.Poll")).Return(nil)

	poll, err := service.CreatePoll(req, "user123")

	assert.Error(t, err)
	assert.Nil(t, poll)
	assert.Contains(t, err.Error(), "question text is required")
	mockRepo.AssertExpectations(t)
}

func TestPollService_CreatePoll_EmptyOptionText(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	req := &model.CreatePollRequest{
		Title:            "Test Poll",
		ResultVisibility: model.ResultVisibilityAlways,
		Questions: []model.QuestionRequest{
			{
				Text:    "Question 1?",
				Options: []string{"Option A", ""},
			},
		},
	}

	mockRepo.On("CreatePoll", mock.Anything, mock.AnythingOfType("*model.Poll")).Return(nil)
	mockRepo.On("CreateQuestion", mock.Anything, mock.AnythingOfType("*model.Question")).Return(nil)
	mockRepo.On("CreateOption", mock.Anything, mock.AnythingOfType("*model.Option")).Return(nil)

	poll, err := service.CreatePoll(req, "user123")

	assert.Error(t, err)
	assert.Nil(t, poll)
	assert.Contains(t, err.Error(), "option text cannot be empty")
	mockRepo.AssertExpectations(t)
}

func TestPollService_CreatePoll_DefaultVisibility(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	req := &model.CreatePollRequest{
		Title: "Test Poll",
		Questions: []model.QuestionRequest{
			{
				Text:    "Question 1?",
				Options: []string{"Option A", "Option B"},
			},
		},
	}

	mockRepo.On("CreatePoll", mock.Anything, mock.AnythingOfType("*model.Poll")).Return(nil)
	mockRepo.On("CreateQuestion", mock.Anything, mock.AnythingOfType("*model.Question")).Return(nil)
	mockRepo.On("CreateOption", mock.Anything, mock.AnythingOfType("*model.Option")).Return(nil)

	poll, err := service.CreatePoll(req, "user123")

	assert.NoError(t, err)
	assert.NotNil(t, poll)
	assert.Equal(t, model.ResultVisibilityAlways, poll.ResultVisibility)
}

func TestPollService_CreatePoll_MultipleQuestions(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	req := &model.CreatePollRequest{
		Title:            "Multi-Question Poll",
		ResultVisibility: model.ResultVisibilityAfterVote,
		Questions: []model.QuestionRequest{
			{
				Text:    "Question 1?",
				Options: []string{"A", "B"},
			},
			{
				Text:    "Question 2?",
				Options: []string{"C", "D", "E"},
			},
		},
	}

	mockRepo.On("CreatePoll", mock.Anything, mock.AnythingOfType("*model.Poll")).Return(nil)
	mockRepo.On("CreateQuestion", mock.Anything, mock.AnythingOfType("*model.Question")).Return(nil)
	mockRepo.On("CreateOption", mock.Anything, mock.AnythingOfType("*model.Option")).Return(nil)

	poll, err := service.CreatePoll(req, "user123")

	assert.NoError(t, err)
	assert.NotNil(t, poll)
	assert.Len(t, poll.Questions, 2)

	mockRepo.AssertExpectations(t)
}

func TestPollService_SubmitVote_Success(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	poll := &model.Poll{
		ID:               "poll-1",
		Title:            "Test Poll",
		ResultVisibility: model.ResultVisibilityAlways,
	}

	mockRepo.On("GetPollByID", mock.Anything, "poll-1").Return(poll, nil)
	mockRepo.On("HasVoted", mock.Anything, "poll-1", "voter-1", "q-1").Return(false, nil)
	mockRepo.On("CreateVote", mock.Anything, mock.AnythingOfType("*model.Vote"), "q-1", "opt-1").Return(nil)

	answers := []model.AnswerRequest{
		{QuestionID: "q-1", OptionID: "opt-1"},
	}

	err := service.SubmitVote("poll-1", "voter-1", answers)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestPollService_SubmitVote_AlreadyVoted(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	poll := &model.Poll{
		ID:               "poll-1",
		Title:            "Test Poll",
		ResultVisibility: model.ResultVisibilityAlways,
	}

	mockRepo.On("GetPollByID", mock.Anything, "poll-1").Return(poll, nil)
	mockRepo.On("HasVoted", mock.Anything, "poll-1", "voter-1", "q-1").Return(true, nil)

	answers := []model.AnswerRequest{
		{QuestionID: "q-1", OptionID: "opt-1"},
	}

	err := service.SubmitVote("poll-1", "voter-1", answers)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already voted")
}

func TestPollService_SubmitVote_MultipleAnswers(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	poll := &model.Poll{
		ID:               "poll-1",
		Title:            "Test Poll",
		ResultVisibility: model.ResultVisibilityAlways,
	}

	mockRepo.On("GetPollByID", mock.Anything, "poll-1").Return(poll, nil)
	mockRepo.On("HasVoted", mock.Anything, "poll-1", "voter-1", "q-1").Return(false, nil)
	mockRepo.On("HasVoted", mock.Anything, "poll-1", "voter-1", "q-2").Return(false, nil)
	mockRepo.On("CreateVote", mock.Anything, mock.AnythingOfType("*model.Vote"), "q-1", "opt-1").Return(nil)
	mockRepo.On("CreateVote", mock.Anything, mock.AnythingOfType("*model.Vote"), "q-2", "opt-3").Return(nil)

	answers := []model.AnswerRequest{
		{QuestionID: "q-1", OptionID: "opt-1"},
		{QuestionID: "q-2", OptionID: "opt-3"},
	}

	err := service.SubmitVote("poll-1", "voter-1", answers)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestPollService_GetPollResults_CanViewAlways(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	poll := &model.Poll{
		ID:               "poll-1",
		Title:            "Test Poll",
		ResultVisibility: model.ResultVisibilityAlways,
		CreatedBy:        "author-1",
	}

	questions := []model.Question{
		{ID: "q-1", PollID: "poll-1", Text: "Question 1?"},
	}

	options := []model.Option{
		{ID: "opt-1", QuestionID: "q-1", Text: "Option A"},
		{ID: "opt-2", QuestionID: "q-1", Text: "Option B"},
	}

	mockRepo.On("GetPollByID", mock.Anything, "poll-1").Return(poll, nil)
	mockRepo.On("GetQuestionsByPollID", mock.Anything, "poll-1").Return(questions, nil)
	mockRepo.On("GetVoteCount", mock.Anything, "poll-1").Return(10, nil)
	mockRepo.On("GetOptionsByQuestionID", mock.Anything, "q-1").Return(options, nil)
	mockRepo.On("GetOptionVoteCount", mock.Anything, "opt-1").Return(6, nil)
	mockRepo.On("GetOptionVoteCount", mock.Anything, "opt-2").Return(4, nil)

	result, err := service.GetPollResults("poll-1", "random-user")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.CanView)
	assert.Equal(t, 10, result.TotalVotes)
	assert.Len(t, result.Poll.Questions, 1)
	assert.Equal(t, 6, result.Poll.Questions[0].Options[0].Votes)
	assert.Equal(t, 4, result.Poll.Questions[0].Options[1].Votes)
}

func TestPollService_GetPollResults_AuthorOnly(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	poll := &model.Poll{
		ID:               "poll-1",
		Title:            "Test Poll",
		ResultVisibility: model.ResultVisibilityAuthorOnly,
		CreatedBy:        "author-1",
	}

	mockRepo.On("GetPollByID", mock.Anything, "poll-1").Return(poll, nil)
	mockRepo.On("GetQuestionsByPollID", mock.Anything, "poll-1").Return([]model.Question{}, nil)
	mockRepo.On("GetVoteCount", mock.Anything, "poll-1").Return(5, nil)

	// Author can view
	result, err := service.GetPollResults("poll-1", "author-1")
	assert.NoError(t, err)
	assert.True(t, result.CanView)

	// Other user cannot view
	result, err = service.GetPollResults("poll-1", "other-user")
	assert.NoError(t, err)
	assert.False(t, result.CanView)
}

func TestPollService_GetPollResults_AfterVote(t *testing.T) {
	// Test with no votes - cannot view
	t.Run("no votes", func(t *testing.T) {
		mockRepo := new(MockPollRepository)
		service := NewPollService(mockRepo)

		poll := &model.Poll{
			ID:               "poll-1",
			Title:            "Test Poll",
			ResultVisibility: model.ResultVisibilityAfterVote,
			CreatedBy:        "author-1",
		}

		mockRepo.On("GetPollByID", mock.Anything, "poll-1").Return(poll, nil)
		mockRepo.On("GetQuestionsByPollID", mock.Anything, "poll-1").Return([]model.Question{}, nil)
		mockRepo.On("GetVoteCount", mock.Anything, "poll-1").Return(0, nil)

		result, err := service.GetPollResults("poll-1", "random-user")
		assert.NoError(t, err)
		assert.False(t, result.CanView)
	})

	// Test with votes - can view
	t.Run("has votes", func(t *testing.T) {
		mockRepo := new(MockPollRepository)
		service := NewPollService(mockRepo)

		poll := &model.Poll{
			ID:               "poll-1",
			Title:            "Test Poll",
			ResultVisibility: model.ResultVisibilityAfterVote,
			CreatedBy:        "author-1",
		}

		mockRepo.On("GetPollByID", mock.Anything, "poll-1").Return(poll, nil)
		mockRepo.On("GetQuestionsByPollID", mock.Anything, "poll-1").Return([]model.Question{}, nil)
		mockRepo.On("GetVoteCount", mock.Anything, "poll-1").Return(5, nil)

		result, err := service.GetPollResults("poll-1", "random-user")
		assert.NoError(t, err)
		assert.True(t, result.CanView)
	})
}

func TestPollService_GetPublicPolls(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	expectedPolls := []model.Poll{
		{ID: "poll-1", Title: "Poll 1", Visibility: model.VisibilityPublic},
		{ID: "poll-2", Title: "Poll 2", Visibility: model.VisibilityPublic},
	}

	mockRepo.On("GetPublicPolls", mock.Anything, 20).Return(expectedPolls, nil)

	polls, err := service.GetPublicPolls(20, "viewer-1")

	assert.NoError(t, err)
	assert.Len(t, polls, 2)
	assert.Equal(t, "Poll 1", polls[0].Title)
	assert.Equal(t, "Poll 2", polls[1].Title)
}

func TestPollService_GetPublicPolls_FiltersUnlisted(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	allPolls := []model.Poll{
		{ID: "poll-1", Title: "Public Poll", Visibility: model.VisibilityPublic, CreatedBy: "author-1"},
		{ID: "poll-2", Title: "Unlisted Poll", Visibility: model.VisibilityUnlisted, CreatedBy: "author-2"},
		{ID: "poll-3", Title: "Another Public", Visibility: model.VisibilityPublic, CreatedBy: "author-3"},
	}

	mockRepo.On("GetPublicPolls", mock.Anything, 20).Return(allPolls, nil)

	// Non-author viewer - should filter out unlisted
	polls, err := service.GetPublicPolls(20, "viewer-1")
	assert.NoError(t, err)
	assert.Len(t, polls, 2)
	assert.Equal(t, "Public Poll", polls[0].Title)
	assert.Equal(t, "Another Public", polls[1].Title)

	// Author of the unlisted poll - should see it
	polls, err = service.GetPublicPolls(20, "author-2")
	assert.NoError(t, err)
	assert.Len(t, polls, 3)
}

func TestPollService_HasVotedForQuestion(t *testing.T) {
	mockRepo := new(MockPollRepository)
	service := NewPollService(mockRepo)

	mockRepo.On("HasVoted", mock.Anything, "poll-1", "voter-1", "q-1").Return(true, nil)

	hasVoted, err := service.HasVotedForQuestion("poll-1", "voter-1", "q-1")

	assert.NoError(t, err)
	assert.True(t, hasVoted)
}
