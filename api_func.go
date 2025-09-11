package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
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

func cleanBody(text string) string {
	profaneWords := []string{"KERFUFFLE", "SHARBERT", "FORNAX"}
	splited := strings.Split(text, " ")
	for i, word := range splited {
		if slices.Contains(profaneWords, strings.ToUpper(word)) {
			splited[i] = "****"
		}
	}
	return strings.Join(splited, " ")
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Body string `json:"body"`
	}

	type resBody struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	req := reqBody{}

	err := decoder.Decode(&req)
	if err != nil {
		errorMessage := "Something went wrong"
		responseError(w, errorMessage, http.StatusBadRequest)
		return
	}

	if len(req.Body) > 140 {
		errorMessage := "Chirp is too long"
		responseError(w, errorMessage, 400)
		return
	}

	cleanedBody := cleanBody(req.Body)

	res := resBody{
		CleanedBody: cleanedBody,
	}

	dat, err := json.Marshal(res)
	if err != nil {
		errorMessage := "Cannot marshal response"
		responseError(w, errorMessage, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)

}

func createUser(cfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type reqBody struct {
			Email string `json:"email"`
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

		user, err := cfg.db.CreateUser(r.Context(), sql.NullString{String: req.Email, Valid: true})
		if err != nil {
			fmt.Println(err)
			errorMessage := "Cannot retrieve user"
			responseError(w, errorMessage, 500)
			return
		}
		res := resBody{
			ID:        user.ID,
			CreatedAt: user.CreatedAt.Time,
			UpdatedAt: user.UpdatedAt.Time,
			Email:     user.Email.String,
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
