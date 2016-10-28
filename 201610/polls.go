package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"regexp"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/pressly/chi"
)

type Poll struct {
	Name     string
	Question string
	Options  []*PollOption
}

type PollOption struct {
	Response string
	Votes    map[string]bool
}

const (
	pollPostMax int64 = 4096
)

var (
	pollCreateTemplate      *template.Template
	pollAddResponseTemplate *template.Template
)

func init() {
	pollCreateTemplate = template.Must(template.ParseFiles("templates/poll-create.html"))
	pollAddResponseTemplate = template.Must(template.ParseFiles("templates/poll-add-response.html"))
}

func PollResponseGet(w http.ResponseWriter, r *http.Request) {
	pollName := chi.URLParam(r, "pollname")
	poll, err := PollByName(pollName)
	if err != nil {
		e := &Error{Code: http.StatusInternalServerError, Message: err}
		e.Write(w, r)
		return
	}
	if poll == nil {
		e := &Error{Code: http.StatusNotFound, Message: errors.New("no such poll")}
		e.Write(w, r)
		return
	}

	err = pollAddResponseTemplate.Execute(w, poll)
	if err != nil {
		e := &Error{
			Code:    http.StatusInternalServerError,
			Message: errors.Wrap(err, "executing poll add response template"),
		}
		e.Write(w, r)
		return
	}
}

func PollsCreateGet(w http.ResponseWriter, r *http.Request) {
	err := pollCreateTemplate.Execute(w, nil)
	if err != nil {
		e := &Error{
			Code:    http.StatusInternalServerError,
			Message: errors.Wrap(err, "executing login template"),
		}
		e.Write(w, r)
		return
	}
}

func PollByName(pollName string) (*Poll, error) {
	var poll *Poll
	err := env.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("polls"))
		if b == nil {
			return errors.New("no poll bucket")
		}
		// iterate over all keys in the bucket
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if string(k) != pollName {
				continue
			}
			poll = &Poll{}
			err := json.Unmarshal(v, poll)
			if err != nil {
				return errors.Wrap(err, "poll unmarshal failed")
			}
		}
		return nil
	})
	return poll, err
}

func AllPolls() (map[string]Poll, error) {
	polls := map[string]Poll{}
	err := env.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("polls"))
		if b == nil {
			return errors.New("no poll bucket")
		}
		// iterate over all keys in the bucket
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var poll Poll
			err := json.Unmarshal(v, &poll)
			if err != nil {
				return errors.Wrap(err, "poll unmarshal failed")
			}
			polls[string(k)] = poll
		}
		return nil
	})
	return polls, err
}

func PollsGet(w http.ResponseWriter, r *http.Request) {
	code := http.StatusInternalServerError
	//inType := r.Context().Value("content-type").(string)
	polls, err := AllPolls()
	if err != nil {
		e := &Error{Code: code, Message: err}
		e.Write(w, r)
		return
	}

	// dump polls to client
	// TODO: format with a template if request wasn't JSON
	if err = json.NewEncoder(w).Encode(polls); err != nil {
		e := &Error{Code: code, Message: errors.Wrap(err, "poll marshal failed")}
		e.Write(w, r)
		return
	}

	return
}

func PollResultsGet(w http.ResponseWriter, r *http.Request) {}

func PollsCreatePost(w http.ResponseWriter, r *http.Request) {
	inType := r.Context().Value("content-type").(string)

	// limit the amount of data we accept for a "poll create" request
	r.Body = http.MaxBytesReader(w, r.Body, pollPostMax)
	defer r.Body.Close()

	var poll *Poll
	var e *Error
	switch {
	case inType == FormURL:
		poll, e = PollFromForm(r)
	case inType == JSON:
		poll, e = PollFromJSON(r.Body)
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

	if e = poll.Save(); e != nil {
		e.Write(w, r)
		return
	}

	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
}

func (p *Poll) Validate() (*Poll, *Error) {
	if len(p.Name) == 0 {
		e := &Error{Code: http.StatusBadRequest, Message: errors.New("Name is required")}
		return nil, e
	} else if len(p.Question) == 0 {
		e := &Error{Code: http.StatusBadRequest, Message: errors.New("Poll is required")}
		return nil, e
	}

	badName, err := regexp.MatchString(`\W`, p.Name)
	if err != nil {
		e := &Error{Code: http.StatusInternalServerError, Message: errors.Wrap(err, "name match failed")}
		return nil, e
	}
	if badName {
		e := &Error{Code: http.StatusBadRequest, Message: errors.New(`poll names must be alphanumeric`)}
		return nil, e
	}

	for _, option := range p.Options {
		if len(option.Response) == 0 {
			e := &Error{Code: http.StatusBadRequest, Message: errors.New("Response is required")}
			return nil, e
		}
	}

	return p, nil
}

func PollFromForm(r *http.Request) (*Poll, *Error) {
	var err error
	poll := &Poll{}

	if err = r.ParseForm(); err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "ParseForm failed"),
		}
		return nil, e
	}

	err = env.Form.Decode(poll, r.PostForm)
	if err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "Decode failed"),
		}
		return nil, e
	}

	return poll.Validate()
}

func PollFromJSON(r io.Reader) (*Poll, *Error) {
	poll := &Poll{}

	if err := json.NewDecoder(r).Decode(poll); err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "json decoding failed"),
		}
		return nil, e
	}

	return poll.Validate()
}

func (p *Poll) Save() *Error {
	code := http.StatusInternalServerError
	jsonBytes, err := json.Marshal(p)
	if err != nil {
		return &Error{
			Code:    http.StatusInternalServerError,
			Message: errors.Wrap(err, "poll marshal failed"),
		}
	}

	err = env.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("polls"))
		if b == nil {
			return errors.New("no poll bucket")
		}
		err := b.Put([]byte(p.Name), jsonBytes)
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

func PollResponsePost(w http.ResponseWriter, r *http.Request) {
	inType := r.Context().Value("content-type").(string)
	pollName := chi.URLParam(r, "pollname")

	// limit the amount of data we accept for a "poll create" request
	r.Body = http.MaxBytesReader(w, r.Body, pollPostMax)
	defer r.Body.Close()

	var option *PollOption
	var e *Error
	switch {
	case inType == FormURL:
		option, e = PollOptionFromForm(r)
	case inType == JSON:
		option, e = PollOptionFromJSON(r.Body)
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

	e = option.Add(pollName)
	if e != nil {
		e.Write(w, r)
		return
	}

	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
}

func (o *PollOption) Add(pollName string) *Error {
	code := http.StatusInternalServerError
	err := env.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("polls"))
		if b == nil {
			code = http.StatusNotFound
			return errors.New("no poll bucket")
		}
		val := b.Get([]byte(pollName))
		if val == nil {
			code = http.StatusNotFound
			return errors.New("no such poll")
		}

		poll := &Poll{}
		if err := json.Unmarshal(val, poll); err != nil {
			return errors.Wrap(err, "poll unmarshal failed")
		}

		for _, option := range poll.Options {
			if option.Response == o.Response {
				return nil
			}
		}
		poll.Options = append(poll.Options, &PollOption{Response: o.Response})

		jsonBytes, err := json.Marshal(poll)
		if err != nil {
			return errors.Wrap(err, "poll marshal failed")
		}
		err = b.Put([]byte(pollName), jsonBytes)
		if err != nil {
			return errors.Wrap(err, "vote failed")
		}

		return nil
	})

	if err != nil {
		return &Error{Code: code, Message: err}
	}
	return nil
}

func PollVotePost(w http.ResponseWriter, r *http.Request) {
	inType := r.Context().Value("content-type").(string)
	pollName := chi.URLParam(r, "pollname")

	// limit the amount of data we accept for a "poll create" request
	r.Body = http.MaxBytesReader(w, r.Body, pollPostMax)
	defer r.Body.Close()

	var option *PollOption
	var e *Error
	switch {
	case inType == FormURL:
		option, e = PollOptionFromForm(r)
	case inType == JSON:
		option, e = PollOptionFromJSON(r.Body)
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

	e = option.Vote(pollName, JWTUser(r))
	if e != nil {
		e.Write(w, r)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/polls/%s", pollName))
	w.WriteHeader(http.StatusFound)
}

func (o *PollOption) Validate() (*PollOption, *Error) {
	if len(o.Response) == 0 {
		e := &Error{Code: http.StatusBadRequest, Message: errors.New("Response is required")}
		return nil, e
	}
	return o, nil
}

func PollOptionFromForm(r *http.Request) (*PollOption, *Error) {
	var err error
	option := &PollOption{}

	if err = r.ParseForm(); err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "ParseForm failed"),
		}
		return nil, e
	}

	err = env.Form.Decode(option, r.PostForm)
	if err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "Decode failed"),
		}
		return nil, e
	}

	return option.Validate()
}

func PollOptionFromJSON(r io.Reader) (*PollOption, *Error) {
	option := &PollOption{}
	if err := json.NewDecoder(r).Decode(option); err != nil {
		e := &Error{
			Code:    http.StatusBadRequest,
			Message: errors.Wrap(err, "json decoding failed"),
		}
		return nil, e
	}
	return option.Validate()
}

func (o *PollOption) Vote(pollName, userName string) *Error {
	code := http.StatusInternalServerError
	err := env.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("polls"))
		if b == nil {
			code = http.StatusNotFound
			return errors.New("no poll bucket")
		}
		val := b.Get([]byte(pollName))
		if val == nil {
			code = http.StatusNotFound
			return errors.New("no such poll")
		}

		poll := &Poll{}
		if err := json.Unmarshal(val, poll); err != nil {
			return errors.Wrap(err, "poll unmarshal failed")
		}
		for _, option := range poll.Options {
			if o.Response == option.Response {
				if option.Votes == nil {
					option.Votes = map[string]bool{}
				}
				option.Votes[userName] = true
			} else {
				delete(option.Votes, userName)
			}
		}

		jsonBytes, err := json.Marshal(poll)
		if err != nil {
			return errors.Wrap(err, "poll marshal failed")
		}
		err = b.Put([]byte(pollName), jsonBytes)
		if err != nil {
			return errors.Wrap(err, "vote failed")
		}

		return nil
	})

	if err != nil {
		return &Error{Code: code, Message: err}
	}
	return nil
}
