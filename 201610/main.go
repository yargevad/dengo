package main

// Process command-line args and initialize listeners

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
)

// Our application-wide configuration.
type Env struct {
	DB *bolt.DB
}

var keyPath = flag.String("keypath", "./.keys", "where to store keys")
var keyName = flag.String("keyname", "201610", "base name for keys")
var port = flag.Int("port", 8080, "where to listen for http requests")
var dbPath = flag.String("dbpath", "bolt.db", "path to db file")

var env = &Env{}

func main() {
	// get values from command line
	flag.Parse()

	router := buildRouter()
	err := env.boltOpen(*dbPath)
	if err != nil {
		log.Fatal(err)
	}

	portSpec := fmt.Sprintf(":%d", *port)
	log.Printf("Listening on %s...\n", portSpec)
	log.Fatal(http.ListenAndServe(portSpec, router))
}
