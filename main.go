package main

import (
	"avito/database"
	"avito/handlers"
	"avito/repository"
	"avito/service"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Дефолтное значение только для локальной разработки
		// В продакшене обязательно используйте переменную окружения DATABASE_URL
		connStr = "host=localhost user=postgres password=postgres dbname=avito sslmode=disable"
		log.Println("Warning: Using default database connection string. Set DATABASE_URL environment variable for production.")
	}

	db, err := database.NewDB(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	repo := repository.NewRepository(db.DB)
	svc := service.NewService(repo)
	h := handlers.NewHandlers(svc)

	r := mux.NewRouter()

	// Team endpoints
	r.HandleFunc("/team/add", h.AddTeam).Methods("POST")
	r.HandleFunc("/team/get", h.GetTeam).Methods("GET")

	// User endpoints
	r.HandleFunc("/users/setIsActive", h.SetUserActive).Methods("POST")

	// PR endpoints
	r.HandleFunc("/pullRequest/create", h.CreatePR).Methods("POST")
	r.HandleFunc("/pullRequest/merge", h.MergePR).Methods("POST")
	r.HandleFunc("/pullRequest/reassign", h.ReassignReviewer).Methods("POST")
	r.HandleFunc("/users/getReview", h.GetReview).Methods("GET")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
