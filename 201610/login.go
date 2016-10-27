package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Name string
	Pass string
}

const (
	loginPostMax int64 = 1024
)

var loginTemplate *template.Template

func init() {
	loginTemplate = template.Must(template.ParseFiles("templates/login.html"))
}

func LoginGet(w http.ResponseWriter, r *http.Request) {
	user := JWTUser(r)
	if user != "" {
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusFound)
		return
	}

	err := loginTemplate.Execute(w, nil)
	if err != nil {
		e := &Error{
			Code:    http.StatusInternalServerError,
			Message: errors.Wrap(err, "executing login template"),
		}
		e.Write(w, r)
		return
	}
}

func LoginPost(w http.ResponseWriter, r *http.Request) {
	inType := r.Context().Value("content-type").(string)

	// limit the amount of data we accept for a login request
	r.Body = http.MaxBytesReader(w, r.Body, loginPostMax)
	defer r.Body.Close()

	var user *User
	var e *Error
	switch {
	case inType == FormURL:
		user, e = UserFromForm(r)
	case inType == JSON:
		user, e = UserFromJSON(r.Body)
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

	e = user.Verify()
	if e != nil {
		e.Write(w, r)
		return
	}

	e = user.SetLoggedIn(w, r)
	if e != nil {
		e.Write(w, r)
		return
	}

	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
}

func (u *User) Validate() (*User, *Error) {
	if len(u.Name) == 0 {
		e := &Error{Code: http.StatusBadRequest, Message: errors.New("Name is required")}
		return nil, e
	} else if len(u.Pass) == 0 {
		e := &Error{Code: http.StatusBadRequest, Message: errors.New("Password is required")}
		return nil, e
	}
	return u, nil
}

func UserFromForm(r *http.Request) (*User, *Error) {
	var err error
	user := &User{}

	if err = r.ParseForm(); err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "ParseForm failed"),
		}
		return nil, e
	}

	err = env.Form.Decode(user, r.PostForm)
	if err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "Decode failed"),
		}
		return nil, e
	}

	return user.Validate()
}

func UserFromJSON(r io.Reader) (*User, *Error) {
	var e *Error
	user := &User{}

	if err := json.NewDecoder(r).Decode(user); err != nil {
		e = &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "json decoding failed"),
		}
		return nil, e
	}

	return user.Validate()
}

func UserByUsername(name string) (*User, *Error) {
	code := http.StatusInternalServerError
	user := &User{}
	err := env.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return errors.New("no such bucket")
		}
		if jsonBytes := b.Get([]byte(name)); jsonBytes != nil {
			err := json.Unmarshal(jsonBytes, user)
			if err != nil {
				return errors.Wrap(err, "user unmarshal failed")
			}
		} else {
			code = http.StatusNotFound
			return errors.New("no such user")
		}
		return nil
	})
	if err != nil {
		e := &Error{Code: code, Message: err}
		return nil, e
	}
	return user, nil
}

func (u *User) Verify() *Error {
	user, e := UserByUsername(u.Name)
	if e != nil {
		return e
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(u.Pass))
	if err != nil {
		e = &Error{
			Code:    http.StatusUnauthorized,
			Message: errors.Wrap(err, "auth failed"),
		}
		return e
	}
	return nil
}

func (u *User) SetLoggedIn(w http.ResponseWriter, r *http.Request) *Error {
	signed, err := JWTString(u.Name)
	if err != nil {
		e := &Error{
			Code:    http.StatusInternalServerError,
			Message: errors.Wrap(err, "jwt encoding failed"),
		}
		return e
	}

	inType := r.Context().Value("content-type").(string)
	switch {
	case inType == FormURL:
		c := &http.Cookie{
			Name: "jwt", Value: signed, HttpOnly: true, // Secure: true,
		}
		http.SetCookie(w, c)
	case inType == JSON:
		w.Header().Set("X-JWT", signed)
	default:
		e := &Error{
			Code:    http.StatusUnsupportedMediaType,
			Message: errors.New("supported types are form, json"),
		}
		return e
	}

	return nil
}

func LogoutGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Set-Cookie", fmt.Sprintf("jwt=; expires=%s; HttpOnly",
		time.Now().Format(time.RFC1123)))
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
}

func LogAuthErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err, ok := ctx.Value("jwt.err").(error); ok {
			e := &Error{Code: http.StatusUnauthorized, Message: err}
			e.Write(w, r)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
