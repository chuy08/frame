package api

import (
	"encoding/json"
	"net/http"
	"time"

	"frame/db"
	"frame/logging"
	"frame/models"

	"go.uber.org/zap"
)

type UserRequest struct {
	Fname string `json:"fname"`
	Lname string `json:"lname"`
}

type UserResponse struct {
	ID        string     `json:"id"`
	FirstName *string    `json:"first_name,omitempty"`
	LastName  *string    `json:"last_name,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// NewUserResponse creates a response struct based on whether the user was created or found
func NewUserResponse(user *models.User, isNewUser bool) UserResponse {
	resp := UserResponse{
		ID: user.ID.String(),
	}

	if isNewUser {
		firstName := user.FirstName
		lastName := user.LastName
		createdAt := user.CreatedAt
		updatedAt := user.UpdatedAt

		resp.FirstName = &firstName
		resp.LastName = &lastName
		resp.CreatedAt = &createdAt
		resp.UpdatedAt = &updatedAt
	}

	return resp
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()

	if r.Method != http.MethodPost {
		logger.Warn("Method not allowed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request",
			zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	logger.Info("Creating new user",
		zap.String("fname", req.Fname),
		zap.String("lname", req.Lname))

	// Create user repository
	userRepo := db.NewUserRepository(db.GetPool())

	// Create user in database
	user, isNewUser, err := userRepo.Create(r.Context(), req.Fname, req.Lname)
	if err != nil {
		logger.Error("Failed to create user",
			zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := NewUserResponse(user, isNewUser)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("Failed to encode response",
			zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
