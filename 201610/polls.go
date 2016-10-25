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

func PollCreate(p *Poll) error {
	return nil
}

func PollListing() ([]Poll, error) {
	return nil, nil
}
