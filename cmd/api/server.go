package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.cfg.port),
		Handler:      app.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  1 * time.Minute,
		ErrorLog:     log.New(app.logger, "", 0),
	}

	shutDownErr := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit
		app.logger.Info("shutting down server", map[string]string{
			"signal": s.String(),
		})
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			shutDownErr <- err
		}

		app.logger.Info("completing background tasks", map[string]string{
			"addr": srv.Addr,
			"env":  app.cfg.env,
		})

		app.wg.Wait()
		shutDownErr <- nil
	}()

	app.logger.Info("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.cfg.env,
	})

	err := srv.ListenAndServe()
	app.wg.Wait()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutDownErr
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", map[string]string{
		"addr": srv.Addr,
		"env":  app.cfg.env,
	})

	return nil
}
