package main

// Process command-line args and initialize listeners

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/boltdb/bolt"
	"github.com/gorilla/schema"
	"github.com/uber-go/zap"
)

// Our application-wide configuration.
type Env struct {
	DB   *bolt.DB
	Log  zap.Logger
	Form *schema.Decoder
}

var keyPath = flag.String("keypath", "./.keys", "where to store keys")
var keyName = flag.String("keyname", "201610", "base name for keys")
var port = flag.Int("port", 8080, "where to listen for http requests")
var dbPath = flag.String("dbpath", "bolt.db", "path to db file")

var env = &Env{}

func main() {
	// get values from command line
	flag.Parse()
	env.Log = zap.New(zap.NewJSONEncoder(), zap.Output(os.Stdout))
	env.Form = schema.NewDecoder()

	router := buildRouter()
	err := env.boltOpen(*dbPath)
	if err != nil {
		env.Log.Fatal(err.Error())
	}

	portSpec := fmt.Sprintf(":%d", *port)
	env.Log.Info("Listening on " + portSpec)
	env.Log.Fatal(http.ListenAndServe(portSpec, router).Error())
}
