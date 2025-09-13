package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/hrncacz/go-chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	dev            bool
	jwtSignString  string
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
	if !cfg.dev {
		w.WriteHeader(403)
		return
	}
	if err := cfg.db.Reset(r.Context()); err != nil {
		w.WriteHeader(500)
		return
	}
	cfg.fileServerHits.Store(0)
	w.WriteHeader(http.StatusOK)
}

func main() {
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatalln(err)
	}
	dbURL := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	dev := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	apiCfg := &apiConfig{
		fileServerHits: atomic.Int32{},
		db:             dbQueries,
		dev:            false,
		jwtSignString:  jwtSecret,
	}
	if dev == "dev" {
		apiCfg.dev = true
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
	mux.HandleFunc("POST /api/users", createUser(apiCfg))
	mux.HandleFunc("GET /api/chirps", getChirpsAll(apiCfg))
	mux.HandleFunc("GET /api/chirps/{chirpID}", getChirpsOne(apiCfg))
	mux.HandleFunc("POST /api/chirps", createChirp(apiCfg))
	mux.HandleFunc("POST /api/login", login(apiCfg))

	defer httpServer.Close()

	if err := httpServer.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
}
