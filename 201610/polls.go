package main

type Poll struct {
	Name     string
	Question string
	Options  []PollOption
}

type PollOption struct {
	Response string
	Votes    map[string]bool
}

func (e *Env) PollCreate(p *Poll) error {
	return nil
}

func (e *Env) PollListing() ([]Poll, error) {
	return nil, nil
}
