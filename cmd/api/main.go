package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"

	"huytran2000-hcmus/greenlight/internal/data"
	"huytran2000-hcmus/greenlight/internal/jsonlog"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	dsn  string
	db   struct {
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rate   float64
		burst  int
		enable bool
	}
}

type application struct {
	logger *jsonlog.Logger
	cfg    config
	models data.Models
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 5000, "API server's port")
	flag.StringVar(&cfg.env, "env", "development", "Enviroment (development|staging|production)")
	flag.StringVar(&cfg.dsn, "dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL data source name")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.Float64Var(&cfg.limiter.rate, "limiter-rate", 2, "Rate limiter average request per seconds")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum request burst")
	flag.BoolVar(&cfg.limiter.enable, "limiter-enable", true, "Rate limiter enable")
	flag.StringVar(
		&cfg.db.maxIdleTime,
		"db-max-idle-time",
		"15m",
		"PostgreSQL max connection idle time",
	)
	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.InfoLevel)

	db, err := openDB(cfg)
	if err != nil {
		logger.FatalErr(err, nil)
	}
	defer db.Close()

	models := data.NewModels(db)

	app := &application{
		logger: logger,
		cfg:    cfg,
		models: models,
	}

	err = app.serve()
	if err != nil {
		logger.FatalErr(err, nil)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %s", err)
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	idleDuration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, fmt.Errorf("max idle time parsing: %s", err)
	}
	db.SetConnMaxIdleTime(idleDuration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("ping db: %s", err)
	}

	return db, nil
}
