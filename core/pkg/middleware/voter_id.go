package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const voterIDKey = "voter_id"

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func VoterID(cookieSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("voter_id")
			var voterID string

			if err != nil || cookie.Value == "" {
				voterID = generateID()
				http.SetCookie(w, &http.Cookie{
					Name:     "voter_id",
					Value:    voterID,
					Path:     "/",
					HttpOnly: true,
					MaxAge:   86400 * 365,
				})
			} else {
				voterID = cookie.Value
			}

			ctx := r.Context()
			ctx = SetVoterID(ctx, voterID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type contextKey string

const voterIDContextKey contextKey = "voter_id"

func SetVoterID(ctx context.Context, voterID string) context.Context {
	return context.WithValue(ctx, voterIDContextKey, voterID)
}

func GetVoterID(r *http.Request) string {
	if voterID, ok := r.Context().Value(voterIDContextKey).(string); ok {
		return voterID
	}
	return ""
}
