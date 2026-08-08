package application

import (
	"context"
	"fmt"
	"net/http"
)

type App struct {
	router http.Handler
}

func NewApp() *App {
	return &App{
		router: routes(),
	}
}

func (a *App) Start(ctx context.Context) error {

	server := &http.Server{
		Addr:    ":8080",
		Handler: a.router,
	}

	err := server.ListenAndServe()

	if err != nil {
		return fmt.Errorf("failed to listen to server: %w", err)
	}

	return nil
}
