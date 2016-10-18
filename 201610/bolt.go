package main

import (
	"time"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

func (e *Env) boltOpen(file string) error {
	var err error
	var opt = bolt.Options{Timeout: 1 * time.Second}
	env.DB, err = bolt.Open(file, 0600, &opt)
	if err != nil {
		return errors.Wrap(err, "boltdb open failed")
	}

	err = env.DB.Update(func(tx *bolt.Tx) error {
		// create buckets (users, polls)
		for _, name := range []string{"users", "polls"} {
			if _, err := tx.CreateBucketIfNotExists([]byte(name)); err != nil {
				return errors.Wrap(err, "boltdb bucket creation failed")
			}
		}
		return nil
	})
	return err
}
