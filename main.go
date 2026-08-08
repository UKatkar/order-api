package main

import (
	"context"
	"fmt"

	"github.com/UKatkar/order-api/application"
)

func main() {

	app := application.NewApp()

	err := app.Start(context.TODO())

	if err != nil {
		fmt.Printf("failed to start the application: %v", err)
	}
}
