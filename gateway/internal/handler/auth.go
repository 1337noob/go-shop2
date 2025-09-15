package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"shop/gateway/internal/middleware"
	"shop/gateway/internal/model"
	"shop/gateway/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	UserID    string `json:"user_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	log.Println(req)

	tx, err := h.db.Begin()
	if err != nil {
		log.Println("Failed to begin transaction", "error", err)
		http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	ctxWithTx := context.WithValue(r.Context(), "tx", tx)

	existUser, err := h.userRepo.FindUserByEmail(ctxWithTx, req.Email)
	if err != nil {
		log.Println("Failed to find user by email", "error", err)
		http.Error(w, "Failed to find user by email", http.StatusInternalServerError)
		return
	}
	if existUser != nil {
		log.Println("User already exists")
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	newUser := &model.User{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	err = h.userRepo.CreateUser(ctxWithTx, newUser)
	if err != nil {
		log.Println("Failed to create user", "error", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	response := RegisterResponse{
		Success: true,
		Message: "Registration successful",
		UserID:  newUser.ID,
	}

	err = tx.Commit()
	if err != nil {
		log.Println("failed to commit transaction", "error", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
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
		Success:   true,
		Message:   "Login successful",
		UserID:    session.UserID,
		SessionID: session.ID,
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
