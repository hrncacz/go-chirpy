package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	httpServer := &http.Server{}

	httpServer.Addr = ":8080"
	httpServer.Handler = mux

	mux.Handle("/", http.FileServer(http.Dir("./")))

	defer httpServer.Close()

	err := httpServer.ListenAndServe()
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
}
