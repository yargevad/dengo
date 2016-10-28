package main

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Signup struct {
	User
	Secret string
}

const (
	signupPostMax int64 = 1024
	bcryptCost    int   = 13
)

var signupTemplate *template.Template

func init() {
	signupTemplate = template.Must(template.ParseFiles("templates/signup.html"))
}

func SignupGet(w http.ResponseWriter, r *http.Request) {
	user := JWTUser(r)
	if user != "" {
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusFound)
		return
	}

	err := signupTemplate.Execute(w, nil)
	if err != nil {
		e := &Error{
			Code:    http.StatusInternalServerError,
			Message: errors.Wrap(err, "executing signup template"),
		}
		e.Write(w, r)
		return
	}
}

func SignupPost(w http.ResponseWriter, r *http.Request) {
	inType := r.Context().Value("content-type").(string)

	// limit the amount of data we accept for a signup request
	r.Body = http.MaxBytesReader(w, r.Body, signupPostMax)
	defer r.Body.Close()

	var signup *Signup
	var e *Error
	switch {
	case inType == FormURL:
		signup, e = SignupFromForm(r)
	case inType == JSON:
		signup, e = SignupFromJSON(r.Body)
	default:
		e = &Error{
			Code:    http.StatusUnsupportedMediaType,
			Message: errors.New("supported types are form, json"),
		}
	}
	if e != nil {
		e.Write(w, r)
		return
	}

	if e = signup.Save(); e != nil {
		e.Write(w, r)
		return
	}

	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
}

func (s *Signup) Validate() (*Signup, *Error) {
	if _, e := s.User.Validate(); e != nil {
		return nil, e
	} else if len(s.Secret) == 0 {
		e := &Error{Code: http.StatusBadRequest, Message: errors.New("Secret is required")}
		return nil, e
	} else if s.Secret != env.Secret {
		e := &Error{Code: http.StatusBadRequest, Message: errors.New("Incorrect secret")}
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
	signup := &Signup{}

	if err := json.NewDecoder(r).Decode(signup); err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "json decoding failed"),
		}
		return nil, e
	}

	return signup.Validate()
}

func (s *Signup) Save() *Error {
	var e *Error
	code := http.StatusInternalServerError
	u := s.User
	// bcrypt password
	bcrypted, err := bcrypt.GenerateFromPassword([]byte(u.Pass), bcryptCost)
	if err != nil {
		e = &Error{
			Code:    http.StatusInternalServerError,
			Message: errors.Wrap(err, "bcrypt failed"),
		}
		return e
	}
	u.Pass = string(bcrypted)
	jsonBytes, err := json.Marshal(u)
	if err != nil {
		e = &Error{
			Code:    http.StatusInternalServerError,
			Message: errors.Wrap(err, "user marshal failed"),
		}
		return e
	}

	err = env.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return errors.New("no users bucket")
		}
		if val := b.Get([]byte(u.Name)); val != nil {
			code = http.StatusConflict
			return errors.New("user exists")
		}
		err := b.Put([]byte(u.Name), jsonBytes)
		if err != nil {
			return errors.Wrap(err, "create failed")
		}
		return nil
	})

	if err != nil {
		return &Error{Code: code, Message: err}
	}
	return nil
}
