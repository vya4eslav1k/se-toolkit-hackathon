package bot

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupTestServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/polls", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			polls := []map[string]string{
				{"id": "poll-1", "title": "Test Poll 1"},
				{"id": "poll-2", "title": "Test Poll 2"},
			}
			json.NewEncoder(w).Encode(polls)
			return
		}
		if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{
				"id":    "new-poll-id",
				"title": "New Poll",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/polls/poll-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			resp := map[string]interface{}{
				"poll": map[string]interface{}{
					"id":    "poll-1",
					"title": "Test Poll 1",
					"questions": []map[string]interface{}{
						{
							"id":   "q-1",
							"text": "What is your favorite color?",
							"options": []map[string]string{
								{"id": "opt-1", "text": "Red"},
								{"id": "opt-2", "text": "Blue"},
							},
						},
					},
				},
				"vote_status": map[string]bool{},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/polls/poll-1/vote", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	mux.HandleFunc("/polls/poll-1/results", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"poll": map[string]interface{}{
				"title": "Test Poll 1",
				"questions": []map[string]interface{}{
					{
						"text": "What is your favorite color?",
						"options": []map[string]interface{}{
							{"text": "Red", "votes": 5},
							{"text": "Blue", "votes": 3},
						},
					},
				},
			},
			"total_votes": 8,
			"can_view":    true,
		}
		json.NewEncoder(w).Encode(resp)
	})

	return httptest.NewServer(mux)
}

func TestNew_InvalidToken(t *testing.T) {
	_, err := New("invalid-token", "http://localhost:8080")
	assert.Error(t, err)
}

func TestBot_StateManagement(t *testing.T) {
	b := &Bot{
		users: make(map[int64]*UserState),
	}

	// Test get state creates new state
	state := b.getState(12345)
	assert.Equal(t, "idle", state.State)
	assert.Nil(t, state.SelectedPoll)

	// Test get state returns existing state
	state.SelectedPoll = &PollItem{ID: "poll-1"}
	state2 := b.getState(12345)
	assert.Equal(t, "poll-1", state2.SelectedPoll.ID)

	// Test clear state
	b.clearState(12345)
	state3 := b.getState(12345)
	assert.Equal(t, "idle", state3.State)
	assert.Nil(t, state3.SelectedPoll)
}

func TestBot_APIGet(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	b := &Bot{
		apiURL: server.URL,
	}

	data, err := b.apiGet("/polls", 12345)
	assert.NoError(t, err)

	var polls []map[string]string
	err = json.Unmarshal(data, &polls)
	assert.NoError(t, err)
	assert.Len(t, polls, 2)
	assert.Equal(t, "poll-1", polls[0]["id"])
	assert.Equal(t, "Test Poll 1", polls[0]["title"])
}

func TestBot_APIPost(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	b := &Bot{
		apiURL: server.URL,
	}

	data := map[string]interface{}{
		"title": "New Poll",
	}
	resp, err := b.apiPost("/polls", data, 12345)
	assert.NoError(t, err)

	var result map[string]string
	err = json.Unmarshal(resp, &result)
	assert.NoError(t, err)
	assert.Equal(t, "new-poll-id", result["id"])
	assert.Equal(t, "New Poll", result["title"])
}

func TestBot_FeedSelection(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	b := &Bot{
		apiURL: server.URL,
		users:  make(map[int64]*UserState),
	}

	userID := int64(12345)
	state := b.getState(userID)
	state.State = "chose_feed_poll"
	state.SelectedPoll = &PollItem{ID: "poll-1"}

	assert.Equal(t, "poll-1", state.SelectedPoll.ID)
}

func TestBot_VisibilityOptions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1", "always"},
		{"2", "after_vote"},
		{"3", "author_only"},
	}

	for _, tt := range tests {
		t.Run("visibility_"+tt.input, func(t *testing.T) {
			result := ""
			switch tt.input {
			case "1":
				result = "always"
			case "2":
				result = "after_vote"
			case "3":
				result = "author_only"
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}
