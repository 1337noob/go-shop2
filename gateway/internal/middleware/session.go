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

		// Обновляем время активности
		session.LastActivity = time.Now()
		session.ExpiresAt = time.Now().Add(m.sessionTTL)

		if err := m.repo.UpdateSession(r.Context(), session); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Добавляем сессию в контекст
		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *SessionMiddleware) OptionalSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.getSessionFromRequest(r)
		if err == nil && session != nil && session.IsValid() {
			// Обновляем сессию если она валидна
			session.LastActivity = time.Now()
			session.ExpiresAt = time.Now().Add(m.sessionTTL)
			m.repo.UpdateSession(r.Context(), session)

			ctx := context.WithValue(r.Context(), "session", session)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

func (m *SessionMiddleware) getSessionFromRequest(r *http.Request) (*model.Session, error) {
	// Пробуем получить сессию из cookie
	cookie, err := r.Cookie(m.sessionName)
	if err == nil && cookie.Value != "" {
		return m.repo.GetSession(r.Context(), cookie.Value)
	}

	// Пробуем получить сессию из заголовка Authorization
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

	// Устанавливаем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     m.sessionName,
		Value:    sessionID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   true, // Только для HTTPS в production
		SameSite: http.SameSiteStrictMode,
	})

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

	// Удаляем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     m.sessionName,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})

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
