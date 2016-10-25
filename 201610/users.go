package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/uber-go/zap"
)

type User struct {
	Name string
	Pass string
}

var signupPostMax int64 = 1024

func SignupPost(w http.ResponseWriter, r *http.Request) {
	var inType string
	var err error
	var e *Error

	// parse incoming content-type
	if inType, err = RequestType(r); err != nil {
		e = &Error{Code: http.StatusBadRequest, Message: err}
		e.Write(w, r)
		return
	}

	// content-type is required
	if inType == "" {
		e = &Error{
			Code:    http.StatusBadRequest,
			Message: errors.New("signup with no content-type"),
		}
		e.Write(w, r)
		return
	}

	// limit the amount of data we accept for a signup request
	r.Body = http.MaxBytesReader(w, r.Body, signupPostMax)
	defer r.Body.Close()

	var user *User
	switch {
	case inType == "application/x-www-form-urlencoded":
		user, e = UserFromForm(r)
		if e != nil {
			e.Write(w, r)
			return
		}
	case inType == "application/json":
		user, e = UserFromJSON(r.Body)
		if e != nil {
			e.Write(w, r)
			return
		}
	default:
		e = &Error{
			Code:    http.StatusUnsupportedMediaType,
			Message: errors.New("supported input types are form, json"),
		}
		e.Write(w, r)
		return
	}
	env.Log.Info("user create", zap.String("name", user.Name))
}

func UserFromForm(r *http.Request) (*User, *Error) {
	var user User

	if err = r.ParseForm(); err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "ParseForm failed"),
		}
		return nil, e
	}

	err = env.Form.Decode(&user, r.PostForm)
	if err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "Decode failed"),
		}
		return nil, e
	}

	return &user, nil
}

func UserFromJSON(r io.Reader) (*User, *Error) {
	var err error
	var user User
	if err = json.NewDecoder(r).Decode(&user); err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "json decoding failed"),
		}
		return nil, e
	}

	return &user, nil
}

func (e *Env) UserCreate(p *Poll) error {
	return nil
}
