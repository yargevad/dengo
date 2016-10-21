package main

import (
	"time"

	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
)

func buildRouter() *chi.Mux {
	r := chi.NewRouter()
	// RequestID is a middleware that injects a request ID into the context of each
	// request. A request ID is a string of the form "host.example.com/random-0001",
	// where "random" is a base62 random string that uniquely identifies this go
	// process, and where the last number is an atomically incremented request
	// counter.
	r.Use(middleware.RequestID)

	// Logger is a middleware that logs the start and end of each request, along
	// with some useful data about what was requested, what the response status was,
	// and how long it took to return. When standard output is a TTY, Logger will
	// print in color, otherwise it will print in black and white.
	//
	// Logger prints a request ID if one is provided.
	r.Use(middleware.Logger)

	// Recoverer is a middleware that recovers from panics, logs the panic (and a
	// backtrace), and returns a HTTP 500 (Internal Server Error) status if
	// possible.
	//
	// Recoverer prints a request ID if one is provided.
	r.Use(middleware.Recoverer)

	// CloseNotify is a middleware that cancels ctx when the underlying
	// connection has gone away. It can be used to cancel long operations
	// on the server when the client disconnects before the response is ready.
	r.Use(middleware.CloseNotify)

	// Timeout is a middleware that cancels ctx after a given timeout and return
	// a 504 Gateway Timeout error to the client.
	//
	// It's required that you select the ctx.Done() channel to check for the signal
	// if the context has reached its deadline and return, otherwise the timeout
	// signal will be just ignored.
	r.Use(middleware.Timeout(10 * time.Second))

	// This application lets users create polls and vote (best beer, best pizza)

	// GETting / shows links to the polls
	//   bonus: with totals cached once a second
	r.Get("/", Index)

	// GETting /login shows auth info form
	// POSTing /login attempts login
	// GETting /logout deletes a user's login cookie(s)
	// GETting /signup shows account info form
	//   no email required, just hardcoded secret from slides
	// POSTing /signup attempts account creation
	// GETting /polls shows paginated list of polls
	// GETting /polls/create shows poll info form
	// GETting /polls/:pollID/results shows poll results

	// POSTing /polls/create attempts poll creation
	// GETting /polls/:pollID displays voting form
	// POSTing /polls/:pollID submits vote

	return r
}
