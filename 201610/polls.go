package main

type Poll struct {
	ID       int
	Question string
	Options  []PollOption
}

type PollOption struct {
	Response string
	Votes    map[string]bool
}

func (e *Env) PollCreate(p *Poll) (int, error) {
	return -1, nil
}

func (e *Env) PollListing() ([]Poll, error) {
	return nil, nil
}
