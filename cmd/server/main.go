package main

import (
	"context"
	"github.com/dlomanov/gophkeeper/cmd/server/config"
	"github.com/dlomanov/gophkeeper/internal/apps/server"
	"log"
)

func main() {
	c := config.Parse()
	if err := server.Run(context.Background(), c); err != nil {
		log.Fatal(err)
	}
}
