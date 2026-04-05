package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/maksimslavik/inno-se-toolkit-pet/core/internal/model"
	"github.com/maksimslavik/inno-se-toolkit-pet/core/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock service
type MockPollService struct {
	mock.Mock
}

func (m *MockPollService) CreatePoll(req *model.CreatePollRequest, createdBy string) (*model.Poll, error) {
	args := m.Called(req, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Poll), args.Error(1)
}

func (m *MockPollService) GetPoll(id string) (*model.Poll, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Poll), args.Error(1)
}

func (m *MockPollService) SubmitVote(pollID, voterID string, answers []model.AnswerRequest) error {
	args := m.Called(pollID, voterID, answers)
	return args.Error(0)
}

func (m *MockPollService) GetPollResults(pollID, viewerID string) (*model.PollResult, error) {
	args := m.Called(pollID, viewerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PollResult), args.Error(1)
}

func (m *MockPollService) GetPublicPolls(limit int, viewerID string) ([]model.Poll, error) {
	args := m.Called(limit, viewerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Poll), args.Error(1)
}

func (m *MockPollService) HasVotedForQuestion(pollID, voterID, questionID string) (bool, error) {
	args := m.Called(pollID, voterID, questionID)
	return args.Bool(0), args.Error(1)
}

func TestPollHandler_CreatePoll_Success(t *testing.T) {
	mockService := new(MockPollService)
	handler := NewPollHandler(mockService)

	reqBody := model.CreatePollRequest{
		Title:            "Test Poll",
		ResultVisibility: model.ResultVisibilityAlways,
		Questions: []model.QuestionRequest{
			{Text: "Q1?", Options: []string{"A", "B"}},
		},
	}

	expectedPoll := &model.Poll{
		ID:               "poll-1",
		Title:            "Test Poll",
		ResultVisibility: model.ResultVisibilityAlways,
	}

	mockService.On("CreatePoll", &reqBody, mock.AnythingOfType("string")).Return(expectedPoll, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/polls", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Add voter ID to context
	ctx := req.Context()
	ctx = middleware.SetVoterID(ctx, "test-voter")
	req = req.WithContext(ctx)

	handler.CreatePoll(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response model.Poll
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "Test Poll", response.Title)

	mockService.AssertExpectations(t)
}

func TestPollHandler_CreatePoll_InvalidBody(t *testing.T) {
	mockService := new(MockPollService)
	handler := NewPollHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/polls", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.CreatePoll(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestPollHandler_GetPoll_Success(t *testing.T) {
	mockService := new(MockPollService)
	handler := NewPollHandler(mockService)

	poll := &model.Poll{
		ID:    "poll-1",
		Title: "Test Poll",
		Questions: []model.Question{
			{ID: "q-1", Text: "Question 1?"},
		},
	}

	mockService.On("GetPoll", "poll-1").Return(poll, nil)
	mockService.On("HasVotedForQuestion", "poll-1", "test-voter", "q-1").Return(false, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/polls/poll-1", nil)

	// Set voter ID using middleware helper
	ctx := middleware.SetVoterID(req.Context(), "test-voter")
	req = req.WithContext(ctx)

	// Add URL param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "poll-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.GetPoll(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NotNil(t, response["poll"])

	mockService.AssertExpectations(t)
}

func TestPollHandler_GetPoll_NotFound(t *testing.T) {
	mockService := new(MockPollService)
	handler := NewPollHandler(mockService)

	mockService.On("GetPoll", "invalid-id").Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/api/polls/invalid-id", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.GetPoll(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestPollHandler_SubmitVote_Success(t *testing.T) {
	mockService := new(MockPollService)
	handler := NewPollHandler(mockService)

	voteReq := model.VoteRequest{
		Answers: []model.AnswerRequest{
			{QuestionID: "q-1", OptionID: "opt-1"},
		},
	}

	mockService.On("SubmitVote", "poll-1", "test-voter", voteReq.Answers).Return(nil)

	body, _ := json.Marshal(voteReq)
	req := httptest.NewRequest(http.MethodPost, "/api/polls/poll-1/vote", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "poll-1")
	ctx := middleware.SetVoterID(req.Context(), "test-voter")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.SubmitVote(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "success", response["status"])

	mockService.AssertExpectations(t)
}

func TestPollHandler_GetResults_Success(t *testing.T) {
	mockService := new(MockPollService)
	handler := NewPollHandler(mockService)

	result := &model.PollResult{
		Poll:       model.Poll{ID: "poll-1", Title: "Test Poll"},
		TotalVotes: 10,
		CanView:    true,
	}

	mockService.On("GetPollResults", "poll-1", "test-voter").Return(result, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/polls/poll-1/results", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "poll-1")
	ctx := middleware.SetVoterID(req.Context(), "test-voter")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.GetResults(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response model.PollResult
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, 10, response.TotalVotes)
	assert.True(t, response.CanView)

	mockService.AssertExpectations(t)
}

func TestPollHandler_GetResults_Forbidden(t *testing.T) {
	mockService := new(MockPollService)
	handler := NewPollHandler(mockService)

	result := &model.PollResult{
		Poll:       model.Poll{ID: "poll-1"},
		TotalVotes: 0,
		CanView:    false,
	}

	mockService.On("GetPollResults", "poll-1", "test-voter").Return(result, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/polls/poll-1/results", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "poll-1")
	ctx := middleware.SetVoterID(req.Context(), "test-voter")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.GetResults(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestPollHandler_GetPublicPolls_Success(t *testing.T) {
	mockService := new(MockPollService)
	handler := NewPollHandler(mockService)

	polls := []model.Poll{
		{ID: "poll-1", Title: "Poll 1"},
		{ID: "poll-2", Title: "Poll 2"},
	}

	mockService.On("GetPublicPolls", 20, "").Return(polls, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/polls", nil)
	rr := httptest.NewRecorder()

	handler.GetPublicPolls(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []model.Poll
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Len(t, response, 2)

	mockService.AssertExpectations(t)
}
