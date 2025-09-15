package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func eventUserUpgraded(cfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type reqBody struct {
			Event string `json:"event"`
			Data  struct {
				UserID uuid.UUID `json:"user_id"`
			} `json:"data"`
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
