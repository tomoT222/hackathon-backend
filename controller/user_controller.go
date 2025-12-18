package controller

import (
	"encoding/json"
	"hackathon-backend/usecase"
	"net/http"
)

type UserController struct {
	usecase *usecase.UserUsecase
}

func NewUserController(usecase *usecase.UserUsecase) *UserController {
	return &UserController{usecase: usecase}
}

type RegisterRequest struct {
	ID    string `json:"id"` // Optional: Firebase UID
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (c *UserController) Register(w http.ResponseWriter, r *http.Request) {
    // CORS headers
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := c.usecase.RegisterUser(req.ID, req.Name, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
