package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"shop/gateway/internal/middleware"
	"shop/gateway/internal/repository"
)

type AuthHandler struct {
	db                *sql.DB
	sessionMiddleware *middleware.SessionMiddleware
	userRepo          repository.UserRepository
}

func NewAuthHandler(db *sql.DB, sessionMiddleware *middleware.SessionMiddleware, userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		db:                db,
		sessionMiddleware: sessionMiddleware,
		userRepo:          userRepo,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	log.Println(req)

	// Здесь должна быть реальная проверка учетных данных
	tx, err := h.db.Begin()
	if err != nil {
		log.Println("Failed to begin transaction", "error", err)
		http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
	}
	defer tx.Rollback()

	ctxWithTx := context.WithValue(r.Context(), "tx", tx)

	user, err := h.userRepo.FindUserByEmail(ctxWithTx, req.Email)
	if err != nil {
		log.Println("Failed to find user by email", "error", err)
		response := LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// TODO check hash password
	if user.Password != req.Password {
		response := LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Создаем сессию
	session, err := h.sessionMiddleware.CreateSession(w, r, user.ID)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	response := LoginResponse{
		Success: true,
		Message: "Login successful",
		UserID:  session.UserID,
	}

	err = tx.Commit()
	if err != nil {
		log.Println("failed to commit transaction", "error", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.sessionMiddleware.DestroySession(w, r); err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Logout successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		log.Println("Failed to begin transaction", "error", err)
		http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
	}
	defer tx.Rollback()

	ctxWithTx := context.WithValue(r.Context(), "tx", tx)

	user, err := h.userRepo.FindUserByID(ctxWithTx, session.UserID)
	if err != nil {
		http.Error(w, "Failed to find user", http.StatusInternalServerError)
		return
	}

	profile := map[string]interface{}{
		"user_id":   user.ID,
		"name":      user.Name,
		"email":     user.Email,
		"logged_in": session.CreatedAt,
	}

	err = tx.Commit()
	if err != nil {
		log.Println("failed to commit transaction", "error", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
