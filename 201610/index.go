package main

import (
	"html/template"
	"net/http"

	"github.com/pkg/errors"
)

type IndexModel struct {
	LoggedIn bool
	Username string
	Polls    map[string]Poll
}

var indexTemplate *template.Template

func init() {
	indexTemplate = template.Must(template.ParseFiles("templates/index.html"))
}

func Index(w http.ResponseWriter, r *http.Request) {
	var err error
	model := &IndexModel{}
	model.Username = JWTUser(r)
	if model.Username != "" {
		model.LoggedIn = true
	}

	model.Polls, err = AllPolls()
	if err != nil {
		e := &Error{Code: http.StatusInternalServerError, Message: err}
		e.Write(w, r)
		return
	}

	err = indexTemplate.Execute(w, model)
	if err != nil {
		e := &Error{
			Code:    http.StatusInternalServerError,
			Message: errors.Wrap(err, "executing index template"),
		}
		e.Write(w, r)
		return
	}
}
