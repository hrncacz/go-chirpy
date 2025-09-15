package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hrncacz/go-chirpy/internal/auth"
	"github.com/hrncacz/go-chirpy/internal/database"
)

func login(cfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type reqBody struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		type resBody struct {
			ID           uuid.UUID `json:"id"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
			Email        string    `json:"email"`
			Token        string    `json:"token"`
			RefreshToken string    `json:"refresh_token"`
		}
		decoder := json.NewDecoder(r.Body)
		req := reqBody{}
		err := decoder.Decode(&req)
		if err != nil {
			errorMessage := "Something went wrong"
			responseError(w, errorMessage, http.StatusBadRequest)
			return
		}
		user, err := cfg.db.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			errorMessage := "Unauthorized"
			responseError(w, errorMessage, 401)
			return
		}
		if err := auth.CheckPasswordHash(req.Password, user.HashedPassword); err != nil {
			errorMessage := "Unauthorized"
			responseError(w, errorMessage, 401)
			return
		}
		jwtTokenExpiration := 1 * time.Hour
		jwtToken, err := auth.MakeJWT(user.ID, cfg.jwtSignString, jwtTokenExpiration)
		if err != nil {
			errorMessage := "JWT issue"
			responseError(w, errorMessage, 500)
			return
		}

		refreshToken, err := auth.MakeRefreshToken()
		if err != nil {
			errorMessage := "Refresh token issue"
			responseError(w, errorMessage, 500)
			return
		}
		_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
			Token:     refreshToken,
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
		})
		if err != nil {
			errorMessage := "Refresh token issue"
			responseError(w, errorMessage, 500)
			return
		}
		res := resBody{
			ID:           user.ID,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			Email:        user.Email,
			Token:        jwtToken,
			RefreshToken: refreshToken,
		}
		data, err := json.Marshal(res)
		if err != nil {
			errorMessage := "Cannot marshal response"
			responseError(w, errorMessage, 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(data)
	}
}

func refresh(cfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type resBody struct {
			Token string `json:"token"`
		}
		refreshToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			errorMessage := "Missing header"
			responseError(w, errorMessage, 401)
			return
		}
		userID, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
		if err != nil {
			errorMessage := "No valid refresh token found"
			responseError(w, errorMessage, 401)
			return
		}
		newJwtToken, err := auth.MakeJWT(userID, cfg.jwtSignString, cfg.jwtExpiration)
		if err != nil {
			errorMessage := "JWT issue"
			responseError(w, errorMessage, 500)
			return
		}
		res := resBody{
			Token: newJwtToken,
		}
		data, err := json.Marshal(res)
		if err != nil {
			errorMessage := "Cannot marshal response"
			responseError(w, errorMessage, 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(data)
	}
}

func revoke(cfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			errorMessage := "Missing header"
			responseError(w, errorMessage, 401)
			return
		}
		_, err = cfg.db.RevokeTokenByToken(r.Context(), refreshToken)
		if err != nil {
			errorMessage := "No valid refresh token found"
			responseError(w, errorMessage, 401)
			return
		}
		w.WriteHeader(204)
	}
}
