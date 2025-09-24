package middleware

import (
	"context"
	"net/http"
	"shop/gateway/internal/model"
	"shop/gateway/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SessionMiddleware struct {
	repo        repository.SessionRepository
	sessionName string
	sessionTTL  time.Duration
}

func NewSessionMiddleware(storage repository.SessionRepository, sessionName string, ttl time.Duration) *SessionMiddleware {
	return &SessionMiddleware{
		repo:        storage,
		sessionName: sessionName,
		sessionTTL:  ttl,
	}
}

func (m *SessionMiddleware) SessionRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.getSessionFromRequest(r)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if session == nil || !session.IsValid() {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		session.LastActivity = time.Now()

		if err := m.repo.UpdateSession(r.Context(), session); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *SessionMiddleware) getSessionFromRequest(r *http.Request) (*model.Session, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		sessionID := strings.TrimPrefix(authHeader, "Bearer ")
		return m.repo.GetSession(r.Context(), sessionID)
	}

	return nil, nil
}

func (m *SessionMiddleware) CreateSession(w http.ResponseWriter, r *http.Request, userID string) (*model.Session, error) {
	sessionID := uuid.New().String()

	session := &model.Session{
		ID:           sessionID,
		UserID:       userID,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(m.sessionTTL),
		UserAgent:    r.UserAgent(),
		IPAddress:    getIPAddress(r),
	}

	if err := m.repo.CreateSession(r.Context(), session); err != nil {
		return nil, err
	}

	return session, nil
}

func (m *SessionMiddleware) DestroySession(w http.ResponseWriter, r *http.Request) error {
	session, err := m.getSessionFromRequest(r)
	if err != nil {
		return err
	}

	if session != nil {
		if err := m.repo.DeleteSession(r.Context(), session.ID); err != nil {
			return err
		}
	}

	return nil
}

func GetSessionFromContext(ctx context.Context) *model.Session {
	if session, ok := ctx.Value("session").(*model.Session); ok {
		return session
	}
	return nil
}

func getIPAddress(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
