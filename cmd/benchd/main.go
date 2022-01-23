package main

import (
	"log"
)

func main() {
	cmdServer := cmdServer{Host: ":2112"}
	log.Println("listening on port", cmdServer.Host)
	if err := cmdServer.Start(); err != nil {
		log.Fatalf("error listening: %v", err)
	}
}
