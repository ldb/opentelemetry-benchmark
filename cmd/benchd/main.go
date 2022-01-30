package main

import (
	"github.com/ldb/openetelemtry-benchmark/command"
	"log"
)

func main() {
	cmdServer := command.Server{Host: ":2112"}
	log.Println("listening on port", cmdServer.Host)
	if err := cmdServer.Start(); err != nil {
		log.Fatalf("error listening: %v", err)
	}
}
