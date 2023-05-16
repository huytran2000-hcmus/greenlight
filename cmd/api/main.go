package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port         int
	env          string
	dsn          string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type application struct {
	logger *log.Logger
	cfg    config
	db     *sql.DB
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 5000, "API server's port")
	flag.StringVar(&cfg.env, "env", "development", "Enviroment (development|staging|production)")
	flag.StringVar(&cfg.dsn, "dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL data source name")
	flag.IntVar(&cfg.maxOpenConns, "db-max-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max idle connections")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags)

	db, err := openDB(cfg)
	if err != nil {
		panic(err)
	}

	app := &application{
		logger: logger,
		cfg:    cfg,
		db:     db,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}

	logger.Printf("Starting the server at %d", cfg.port)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.maxOpenConns)
	db.SetMaxIdleConns(cfg.maxIdleConns)

	idleDuration, err := time.ParseDuration(cfg.maxIdleTime)
	db.SetConnMaxIdleTime(idleDuration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
