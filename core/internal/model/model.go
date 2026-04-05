package model

import "time"

type Visibility string

const (
	VisibilityPublic   Visibility = "public"
	VisibilityUnlisted Visibility = "unlisted"
)

type ResultVisibility string

const (
	ResultVisibilityAfterVote  ResultVisibility = "after_vote"
	ResultVisibilityAlways     ResultVisibility = "always"
	ResultVisibilityAuthorOnly ResultVisibility = "author_only"
)

type Poll struct {
	ID               string           `json:"id"`
	Title            string           `json:"title"`
	CreatedAt        time.Time        `json:"created_at"`
	CreatedBy        string           `json:"created_by"`
	Visibility       Visibility       `json:"visibility"`
	ResultVisibility ResultVisibility `json:"result_visibility"`
	Questions        []Question       `json:"questions,omitempty"`
}

type Question struct {
	ID      string   `json:"id"`
	PollID  string   `json:"poll_id"`
	Text    string   `json:"text"`
	Options []Option `json:"options"`
}

type Option struct {
	ID         string `json:"id"`
	QuestionID string `json:"question_id"`
	Text       string `json:"text"`
	Votes      int    `json:"votes,omitempty"`
}

type Vote struct {
	ID        string    `json:"id"`
	PollID    string    `json:"poll_id"`
	VoterID   string    `json:"voter_id"`
	CreatedAt time.Time `json:"created_at"`
}

type VoteRequest struct {
	Answers []AnswerRequest `json:"answers" binding:"required"`
}

type AnswerRequest struct {
	QuestionID string `json:"question_id" binding:"required"`
	OptionID   string `json:"option_id" binding:"required"`
}

type CreatePollRequest struct {
	Title            string            `json:"title" binding:"required"`
	Visibility       Visibility        `json:"visibility"`
	ResultVisibility ResultVisibility  `json:"result_visibility"`
	Questions        []QuestionRequest `json:"questions" binding:"required,min=1"`
}

type QuestionRequest struct {
	Text    string   `json:"text" binding:"required"`
	Options []string `json:"options" binding:"required,min=2"`
}

type PollResult struct {
	Poll       Poll `json:"poll"`
	TotalVotes int  `json:"total_votes"`
	CanView    bool `json:"can_view"`
}
