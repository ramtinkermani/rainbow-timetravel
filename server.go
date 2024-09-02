package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rainbowmga/timetravel/api"
	"github.com/rainbowmga/timetravel/storageServices"
)

// logError logs all non-nil errors
func logError(err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func main() {
	// Get preferred storage type from user through CLI args or use default
	storageType := storageServices.GetStorageType()

	// Based on the storage type choice, create the right storage service
	storageService, db := storageServices.BuildStorageService(storageType)
	
	// If db is not nil, defer its closure
	if db != nil {
		defer db.Close()
	}

	router := mux.NewRouter()

	api := api.NewAPI(storageService)

	apiRoute := router.PathPrefix("/api/v1").Subrouter()
	apiRoute.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		logError(err)
	})
	api.CreateRoutes(apiRoute)

	// Setting up routes for API version 2
	apiV2Route := router.PathPrefix("/api/v2").Subrouter()
	apiV2Route.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
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
