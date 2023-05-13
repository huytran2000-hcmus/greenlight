package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
}

type application struct {
	log *log.Logger
	cfg config
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 5000, "API server's port")
	flag.StringVar(&cfg.env, "env", "development", "Enviroment (development|staging|production)")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags)

	app := &application{
		log: logger,
		cfg: cfg,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}

	logger.Printf("Starting the server at %d", cfg.port)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}
