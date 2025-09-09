package main

import (
	"encoding/json"
	"log"
	"net/http"
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

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Body string `json:"body"`
	}

	type resBody struct {
		Valid bool `json:"valid"`
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

	res := resBody{
		Valid: true,
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
