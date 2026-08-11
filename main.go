package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/UKatkar/order-api/application"
)

func main() {

	app := application.NewApp()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	err := app.Start(ctx)

	if err != nil {
		fmt.Printf("failed to start the application: %v", err)
	}
}
