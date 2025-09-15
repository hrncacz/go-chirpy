package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hrncacz/go-chirpy/internal/auth"
	"github.com/hrncacz/go-chirpy/internal/database"
)

func createUser(cfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type reqBody struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		type resBody struct {
			ID        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Email     string    `json:"email"`
		}

		decoder := json.NewDecoder(r.Body)
		req := reqBody{}

		err := decoder.Decode(&req)
		if err != nil {
			fmt.Println(err)
			errorMessage := "Something went wrong"
			responseError(w, errorMessage, http.StatusBadRequest)
			return
		}

		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			fmt.Println(err)
			errorMessage := "Cannot create user"
			responseError(w, errorMessage, 500)
			return
		}

		user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
			Email:          req.Email,
			HashedPassword: hashedPassword,
		})
		if err != nil {
			fmt.Println(err)
			errorMessage := "Cannot retrieve user"
			responseError(w, errorMessage, 500)
			return
		}
		res := resBody{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		}
		data, err := json.Marshal(res)
		if err != nil {
			errorMessage := "Cannot marshal response"
			responseError(w, errorMessage, 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		w.Write(data)
	}

}

func changeEmailPassword(cfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type reqBody struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		accessToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			errorMessage := "Missing header"
			responseError(w, errorMessage, 401)
			return
		}
		userID, err := auth.ValidateJWT(accessToken, cfg.jwtSignString)
		if err != nil {
			errorMessage := "Invalid JWT"
			responseError(w, errorMessage, 401)
			return
		}
		decoder := json.NewDecoder(r.Body)
		req := reqBody{}
		if err = decoder.Decode(&req); err != nil {
			errorMessage := "Invalid body"
			responseError(w, errorMessage, 401)
			return
		}
		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			errorMessage := "Invalid password"
			responseError(w, errorMessage, 500)
			return
		}
		user, err := cfg.db.UpdateUsersEmailPassword(r.Context(), database.UpdateUsersEmailPasswordParams{
			ID:             userID,
			Email:          req.Email,
			HashedPassword: hashedPassword,
		})
		if err != nil {
			errorMessage := "User not found"
			responseError(w, errorMessage, 401)
			return
		}
		data, err := json.Marshal(user)
		if err != nil {
			errorMessage := "Unable to marshal response data"
			responseError(w, errorMessage, 401)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(data)

	}
}
