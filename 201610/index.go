package main

import (
	"fmt"
	"net/http"

	_ "github.com/pkg/errors"
)

func Index(w http.ResponseWriter, r *http.Request) {
	polls, err := env.PollListing()
	if err != nil {
		e := &Error{500, err}
		e.Write(w, r)
	}

	//e := &Error{403, errors.New("foo")}
	//e.Write(w, r)

	fmt.Fprintf(w, "%v+", polls)
}
