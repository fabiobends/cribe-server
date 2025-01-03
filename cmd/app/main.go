package main

import (
	"log"

	router "cribeapp.com/cribe-server/internal/core"
	"cribeapp.com/cribe-server/internal/utils"
)

func main() {
	port := utils.GetPort()

	log.Printf("Listening on port %s", port)

	err := router.Handler(port)

	if err != nil {
		log.Fatal(err)
	}
}
