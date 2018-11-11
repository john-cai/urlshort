package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/john-cai/urlshort/server"
)

func main() {
	s := server.NewServer()
	httpServer := http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler: s,
	}
	log.Fatal(httpServer.ListenAndServe())
}
