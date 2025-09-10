package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
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
