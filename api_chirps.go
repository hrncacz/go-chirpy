package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/google/uuid"
	"github.com/hrncacz/go-chirpy/internal/auth"
	"github.com/hrncacz/go-chirpy/internal/database"
)

func getChirpsAll(cfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorID := r.URL.Query().Get("author_id")
		sortChirps := r.URL.Query().Get("sort")
		var chirps []database.Chirp
		var err error
		if len(authorID) > 0 {
			authorUUID, err := uuid.Parse(authorID)
			if err != nil {
				fmt.Println(err)
				errorMessage := "Invalid authorID"
				responseError(w, errorMessage, 401)
				return
			}
			chirps, err = cfg.db.GetChirpsAllByUserID(r.Context(), authorUUID)
			if err != nil {
				fmt.Println(err)
				errorMessage := "No chirps found"
				responseError(w, errorMessage, 401)
				return
			}
		} else {
			chirps, err = cfg.db.GetChirpsAll(r.Context())
			if err != nil {
				fmt.Println(err)
				errorMessage := "Cannot retrieve chirps"
				responseError(w, errorMessage, 500)
				return
			}
		}
		if sortChirps == "desc" {
			sort.Slice(chirps, func(a, b int) bool { return chirps[a].CreatedAt.After(chirps[b].CreatedAt) })
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
			Body string `json:"body"`
		}

		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			errorMessage := "Unauthorized"
			responseError(w, errorMessage, 401)
			return
		}
		userID, err := auth.ValidateJWT(token, cfg.jwtSignString)
		if err != nil {
			errorMessage := "Unauthorized"
			responseError(w, errorMessage, 401)
			return
		}

		decoder := json.NewDecoder(r.Body)
		req := reqBody{}
		err = decoder.Decode(&req)
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
			UserID: userID,
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

func deleteChirp(cfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			errorMessage := "Unauthorized"
			responseError(w, errorMessage, 401)
			return
		}
		userID, err := auth.ValidateJWT(token, cfg.jwtSignString)
		if err != nil {
			errorMessage := "Unauthorized"
			responseError(w, errorMessage, 401)
			return
		}

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
		if userID != chirps.UserID {
			errorMessage := "user is not owner of chirp"
			responseError(w, errorMessage, 403)
			return
		}
		if err = cfg.db.DeleteChirpById(r.Context(), database.DeleteChirpByIdParams{
			ID:     chirpID,
			UserID: userID,
		}); err != nil {
			errorMessage := fmt.Sprintf("Chirp was not deleted: %s", chirpIDString)
			responseError(w, errorMessage, 404)
			return
		}
		w.WriteHeader(204)
	}
}
