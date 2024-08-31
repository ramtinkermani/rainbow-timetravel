package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rainbowmga/timetravel/api"
	"github.com/rainbowmga/timetravel/service"
)

// logError logs all non-nil errors
func logError(err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func initDB() (*sql.DB, error) {
    db, err := sql.Open("sqlite3", "./data.db")
    if err != nil {
        return nil, err
    }

    createTableQuery := `
    CREATE TABLE IF NOT EXISTS CustomerData (
        id INTEGER PRIMARY KEY,
        data TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT NULL
    );`

    _, err = db.Exec(createTableQuery)
    if err != nil {
        return nil, err
    }

    return db, nil
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatal((err))
	}
	defer db.Close()

	router := mux.NewRouter()

	// service := service.NewInMemoryRecordService()
	service := service.NewSqliteRecordService(db)
	api := api.NewAPI(&service)

	apiRoute := router.PathPrefix("/api/v1").Subrouter()
	apiRoute.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		logError(err)
	})
	api.CreateRoutes(apiRoute)

	// Setting up routes for API version 2
	apiV2Route := router.PathPrefix("/api/v2").Subrouter()
	apiV2Route.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"okkkk": true})
		logError(err)
	})

	api.CreateRoutes(apiV2Route)

	address := "127.0.0.1:8000"
	srv := &http.Server{
		Handler:      router,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("listening on %s", address)
	log.Fatal(srv.ListenAndServe())
}
