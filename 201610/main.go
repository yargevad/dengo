package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var port = flag.Int("port", 8080, "where to listen for http requests")

func main() {
	// get values from command line
	flag.Parse()

	router := buildRouter()
	portSpec := fmt.Sprintf(":%d", *port)
	log.Printf("Listening on %s...\n", portSpec)
	log.Fatal(http.ListenAndServe(portSpec, r))
}
