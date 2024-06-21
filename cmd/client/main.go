package main

import (
	"context"
	"github.com/dlomanov/gophkeeper/cmd/client/config"
	"github.com/dlomanov/gophkeeper/internal/apps/client"
	"log"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	c := config.Parse(false)
	c.BuildVersion = buildVersion
	c.BuildDate = buildDate
	c.BuildCommit = buildCommit
	if err := client.Run(context.Background(), &c); err != nil {
		log.Fatal(err)
	}
}
