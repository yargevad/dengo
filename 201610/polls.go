package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type Poll struct {
	Name     string
	Question string
	Options  []PollOption
}

type PollOption struct {
	Response string
	Votes    map[string]bool
}

const (
	pollPostMax int64 = 4096
)

func PollsGet(w http.ResponseWriter, r *http.Request)       {}
func PollResultsGet(w http.ResponseWriter, r *http.Request) {}
func PollsCreateGet(w http.ResponseWriter, r *http.Request) {}

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
}

func (p *Poll) Validate() (*Poll, *Error) {
	if len(p.Name) == 0 {
		e := &Error{Code: http.StatusBadRequest, Message: errors.New("Name is required")}
		return nil, e
	} else if len(p.Question) == 0 {
		e := &Error{Code: http.StatusBadRequest, Message: errors.New("Poll is required")}
		return nil, e
	} else if len(p.Options) == 0 {
		e := &Error{Code: http.StatusBadRequest, Message: errors.New("Options is required")}
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

func (p *Poll) Save() *Error { return nil }

func PollResponsePost(w http.ResponseWriter, r *http.Request) {}
func PollVoteGet(w http.ResponseWriter, r *http.Request)      {}
func PollVotePost(w http.ResponseWriter, r *http.Request)     {}

func PollCreate(p *Poll) error {
	return nil
}

func PollListing() ([]Poll, error) {
	return nil, nil
}
