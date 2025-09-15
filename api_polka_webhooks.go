package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/hrncacz/go-chirpy/internal/auth"
)

func eventUserUpgraded(cfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type reqBody struct {
			Event string `json:"event"`
			Data  struct {
				UserID uuid.UUID `json:"user_id"`
			} `json:"data"`
		}
		apiKey, err := auth.GetAPIKey(r.Header)
		if err != nil {
			errorMessage := "Invalid header"
			responseError(w, errorMessage, 401)
			return
		}
		if apiKey != cfg.polkaAPIKey {
			errorMessage := "Invalid API key"
			responseError(w, errorMessage, 401)
			return
		}
		decoder := json.NewDecoder(r.Body)
		req := reqBody{}
		if err := decoder.Decode(&req); err != nil {
			errorMessage := "Unable to decode body"
			responseError(w, errorMessage, 500)
			return
		}
		if req.Event != "user.upgraded" {
			w.WriteHeader(204)
			return
		}
		if err := cfg.db.SetIsChirpyRed(r.Context(), req.Data.UserID); err != nil {
			w.WriteHeader(404)
			return
		}
		w.WriteHeader(204)
	}
}
