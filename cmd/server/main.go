package main

import (
	"github.com/dlomanov/gophkeeper/cmd/server/config"
	"github.com/dlomanov/gophkeeper/internal/apps/server"
	"log"
)

func main() {
	c := config.Parse()
	if err := server.Run(c); err != nil {
		log.Fatal(err)
	}
}
