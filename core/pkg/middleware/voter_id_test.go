package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVoterID_NewCookie(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		voterID := GetVoterID(r)
		assert.NotEmpty(t, voterID)
		w.WriteHeader(http.StatusOK)
	})

	middleware := VoterID("test-secret")(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Check cookie was set
	cookies := rr.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "voter_id", cookies[0].Name)
	assert.NotEmpty(t, cookies[0].Value)
	assert.True(t, cookies[0].HttpOnly)
}

func TestVoterID_ExistingCookie(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		voterID := GetVoterID(r)
		assert.Equal(t, "existing-voter-id", voterID)
		w.WriteHeader(http.StatusOK)
	})

	middleware := VoterID("test-secret")(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "voter_id",
		Value: "existing-voter-id",
	})

	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetVoterID_NoCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	voterID := GetVoterID(req)
	assert.Empty(t, voterID)
}

func TestSetVoterID_GetVoterID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := SetVoterID(req.Context(), "test-voter-123")
	req = req.WithContext(ctx)

	voterID := GetVoterID(req)
	assert.Equal(t, "test-voter-123", voterID)
}
