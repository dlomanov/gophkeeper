package main

import (
	"fmt"
	main2 "github.com/dlomanov/gophkeeper/cmd/server/config"
)

func main() {
	_ = main2.Parse()
	fmt.Println("this is client")
}
