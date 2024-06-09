package main

import (
	"context"
	"github.com/dlomanov/gophkeeper/cmd/client/config"
	"github.com/dlomanov/gophkeeper/internal/apps/client"
	"log"
)

func main() {
	c := config.Parse(false)
	if err := client.Run(context.Background(), &c); err != nil {
		log.Fatal(err)
	}
}
