package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMeticsLog(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	nrHits := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`, cfg.fileServerHits.Load())

	w.Write([]byte(nrHits))
}

func (cfg *apiConfig) middlewareMeticsReset(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	cfg.fileServerHits.Store(0)
}

func main() {
	apiCfg := &apiConfig{
		fileServerHits: atomic.Int32{},
	}
	mux := http.NewServeMux()
	httpServer := &http.Server{}

	httpServer.Addr = ":8080"
	httpServer.Handler = mux
	//APP
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("./app/")))))
	//ADMIN
	mux.HandleFunc("GET /admin/metrics", apiCfg.middlewareMeticsLog)
	mux.HandleFunc("POST /admin/reset", apiCfg.middlewareMeticsReset)
	//API
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

	defer httpServer.Close()

	err := httpServer.ListenAndServe()
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
}
