package main

import (
	"fmt"

	"github.com/dlomanov/gophkeeper/cmd/server/config"
)

func main() {
	_ = config.Parse()
	fmt.Println("this is server")
}
