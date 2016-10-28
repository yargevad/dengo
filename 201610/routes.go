package main

import (
	"net/http"
	"time"

	"github.com/goware/jwtauth"
	"github.com/pressly/chi"
	"github.com/yargevad/chi/middleware"
)

func buildRouter() http.Handler {
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
	//
	// ZapLogger is an instance of Logger customized to format errors using zap
	r.Use(ZapLogger)

	// ContentTypeChecks is a middleware that asserts Content-Type is set for POSTs
	r.Use(ContentTypeChecks)

	// Recoverer is a middleware that recovers from panics, logs the panic (and a
	// backtrace), and returns a HTTP 500 (Internal Server Error) status if
	// possible.
	//
	// Recoverer prints a request ID if one is provided.
	//
	// ZapRecoverer is an instance of Recoverer customized to format errors using zap
	r.Use(ZapRecoverer)

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

	r.Use(tokenAuth.Verifier)
	// GETting / shows links to the polls
	//   bonus: with totals cached once a second
	r.Get("/", Index)

	// GETting /login shows auth info form
	r.Get("/login", LoginGet)
	// Attempts login
	r.Post("/login", LoginPost)

	// Deletes a user's login cookie(s)
	r.Get("/logout", LogoutGet)
	// GETting /signup shows account info form
	//   no email required, just hardcoded secret from slides
	r.Get("/signup", SignupGet)
	// Attempts account creation
	r.Post("/signup", SignupPost)

	r.Route("/polls", func(r chi.Router) {
		// Shows paginated list of polls
		r.Get("/", PollsGet)
		// Shows poll results
		r.Get("/:pollname/results", PollResultsGet)

		// The handlers in this group reqire successful login first.
		r.Group(func(r chi.Router) {
			r.Use(LogAuthErrors)
			r.Use(jwtauth.Authenticator)

			// Shows poll info form
			r.Get("/create", PollsCreateGet)
			// Attempts poll creation
			r.Post("/create", PollsCreatePost)
			r.Get("/:pollname/response", PollResponseGet)
			// Adds a response to an existing poll
			r.Post("/:pollname/response", PollResponsePost)
			// Displays voting/status form
			r.Get("/:pollname", PollViewGet)
			// Submits vote
			r.Post("/:pollname", PollVotePost)
		})
	})

	return r
}
