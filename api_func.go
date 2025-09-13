package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hrncacz/go-chirpy/internal/auth"
	"github.com/hrncacz/go-chirpy/internal/database"
)

func responseError(w http.ResponseWriter, errorMessage string, code int) {
	type resError struct {
		Error string `json:"error"`
	}

	res := resError{
		Error: errorMessage,
	}

	dat, err := json.Marshal(res)
	if err != nil {
		log.Printf("Error marshalling error: %s\n", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

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

func getChirpsAll(cfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chirps, err := cfg.db.GetChirpsAll(r.Context())
		if err != nil {
			fmt.Println(err)
			errorMessage := "Cannot retrieve chirps"
			responseError(w, errorMessage, 500)
			return
		}
		data, err := json.Marshal(chirps)
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

func getChirpsOne(cfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chirpIDString := r.PathValue("chirpID")
		chirpID, err := uuid.Parse(chirpIDString)
		if err != nil {
			errorMessage := "Invalid chirp ID"
			responseError(w, errorMessage, 400)
			return
		}
		chirps, err := cfg.db.GetChirpsOne(r.Context(), chirpID)
		if err != nil {
			errorMessage := fmt.Sprintf("Chirp was not found: %s", chirpIDString)
			responseError(w, errorMessage, 404)
			return
		}
		data, err := json.Marshal(chirps)
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

func createChirp(cfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type reqBody struct {
			Body   string    `json:"body"`
			UserID uuid.UUID `json:"user_id"`
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

		if len(req.Body) > 140 {
			errorMessage := "Chirp is too long"
			responseError(w, errorMessage, 400)
			return
		}

		chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
			Body:   req.Body,
			UserID: req.UserID,
		})
		if err != nil {
			fmt.Println(err)
			errorMessage := "Cannot create chirp"
			responseError(w, errorMessage, 500)
			return
		}
		data, err := json.Marshal(chirp)
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

func login(cfg *apiConfig) http.HandlerFunc {
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
		w.WriteHeader(200)
		w.Write(data)
	}
}
