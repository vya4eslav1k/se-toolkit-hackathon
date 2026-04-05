package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/maksimslavik/inno-se-toolkit-pet/core/internal/model"
	"github.com/maksimslavik/inno-se-toolkit-pet/core/pkg/middleware"
)

type PollService interface {
	CreatePoll(req *model.CreatePollRequest, createdBy string) (*model.Poll, error)
	GetPoll(id string) (*model.Poll, error)
	SubmitVote(pollID, voterID string, answers []model.AnswerRequest) error
	GetPollResults(pollID, viewerID string) (*model.PollResult, error)
	GetPublicPolls(limit int, viewerID string) ([]model.Poll, error)
	HasVotedForQuestion(pollID, voterID, questionID string) (bool, error)
}

type PollHandler struct {
	service PollService
}

func NewPollHandler(service PollService) *PollHandler {
	return &PollHandler{service: service}
}

func (h *PollHandler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	createdBy := middleware.GetVoterID(r)

	poll, err := h.service.CreatePoll(&req, createdBy)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, poll)
}

func (h *PollHandler) GetPoll(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "poll ID is required")
		return
	}

	poll, err := h.service.GetPoll(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "poll not found")
		return
	}

	voterID := middleware.GetVoterID(r)
	voteStatus := make(map[string]bool)
	for _, q := range poll.Questions {
		hasVoted, _ := h.service.HasVotedForQuestion(id, voterID, q.ID)
		voteStatus[q.ID] = hasVoted
	}

	response := map[string]interface{}{
		"poll":        poll,
		"vote_status": voteStatus,
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *PollHandler) SubmitVote(w http.ResponseWriter, r *http.Request) {
	pollID := chi.URLParam(r, "id")
	if pollID == "" {
		writeError(w, http.StatusBadRequest, "poll ID is required")
		return
	}

	var req model.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	voterID := middleware.GetVoterID(r)

	if err := h.service.SubmitVote(pollID, voterID, req.Answers); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *PollHandler) GetResults(w http.ResponseWriter, r *http.Request) {
	pollID := chi.URLParam(r, "id")
	if pollID == "" {
		writeError(w, http.StatusBadRequest, "poll ID is required")
		return
	}

	voterID := middleware.GetVoterID(r)

	result, err := h.service.GetPollResults(pollID, voterID)
	if err != nil {
		writeError(w, http.StatusNotFound, "poll not found")
		return
	}

	if !result.CanView {
		writeError(w, http.StatusForbidden, "results not available. Vote first or this poll is author-only.")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *PollHandler) GetPublicPolls(w http.ResponseWriter, r *http.Request) {
	limit := 20
	voterID := middleware.GetVoterID(r)
	polls, err := h.service.GetPublicPolls(limit, voterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch polls")
		return
	}

	if polls == nil {
		polls = []model.Poll{}
	}

	writeJSON(w, http.StatusOK, polls)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
