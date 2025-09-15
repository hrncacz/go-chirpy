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
