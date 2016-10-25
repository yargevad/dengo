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

type Signup struct {
	User

	Secret string
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

	var signup *Signup
	switch {
	case inType == "application/x-www-form-urlencoded":
		signup, e = SignupFromForm(r)
		if e != nil {
			e.Write(w, r)
			return
		}
	case inType == "application/json":
		signup, e = SignupFromJSON(r.Body)
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
	env.Log.Info("signup", zap.String("name", signup.Name))
}

func (s *Signup) Validate() (*Signup, *Error) {
	if len(s.Name) == 0 {
		e := &Error{Code: http.StatusBadRequest, Message: errors.New("Name is required")}
		return nil, e
	} else if len(s.Pass) == 0 {
		e := &Error{Code: http.StatusBadRequest, Message: errors.New("Password is required")}
		return nil, e
	} else if len(s.Secret) == 0 {
		e := &Error{Code: http.StatusBadRequest, Message: errors.New("Secret is required")}
		return nil, e
	}
	return s, nil
}

func SignupFromForm(r *http.Request) (*Signup, *Error) {
	var err error
	signup := &Signup{}

	if err = r.ParseForm(); err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "ParseForm failed"),
		}
		return nil, e
	}

	err = env.Form.Decode(signup, r.PostForm)
	if err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "Decode failed"),
		}
		return nil, e
	}

	return signup.Validate()
}

func SignupFromJSON(r io.Reader) (*Signup, *Error) {
	var err error
	signup := &Signup{}

	if err = json.NewDecoder(r).Decode(signup); err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "json decoding failed"),
		}
		return nil, e
	}

	return signup.Validate()
}

func (e *Env) UserCreate(p *Poll) error {
	return nil
}
