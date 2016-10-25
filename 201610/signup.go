package main

import (
	"encoding/json"
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

func SignupPost(w http.ResponseWriter, r *http.Request) {
	var inType string
	var e *Error

	inType = r.Context().Value("content-type").(string)

	// limit the amount of data we accept for a signup request
	r.Body = http.MaxBytesReader(w, r.Body, signupPostMax)
	defer r.Body.Close()

	var signup *Signup
	switch {
	case inType == ctypeURLForm:
		signup, e = SignupFromForm(r)
	case inType == ctypeJSON:
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

	if e = UserCreate(&signup.User); e != nil {
		e.Write(w, r)
		return
	}
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
	var e *Error
	signup := &Signup{}

	if err := json.NewDecoder(r).Decode(signup); err != nil {
		e = &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "json decoding failed"),
		}
		return nil, e
	}

	return signup.Validate()
}

func UserCreate(u *User) *Error {
	var e *Error
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
	json, err := json.Marshal(u)
	if err != nil {
		e = &Error{
			Code:    http.StatusInternalServerError,
			Message: errors.Wrap(err, "user marshal failed"),
		}
		return e
	}

	err = env.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return errors.Wrap(err, "bucket error")
		}
		if val := b.Get([]byte(u.Name)); val != nil {
			return errors.New("user exists")
		}
		err = b.Put([]byte(u.Name), json)
		if err != nil {
			return errors.Wrap(err, "create failed")
		}
		err = tx.Commit()
		if err != nil {
			return errors.Wrap(err, "commit failed")
		}
		return nil
	})

	if err != nil {
		if err.Error() == "user exists" {
			e = &Error{Code: http.StatusConflict, Message: err}
		} else {
			e = &Error{Code: http.StatusInternalServerError, Message: err}
		}
		return e
	}
	return nil
}
