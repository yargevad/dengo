// Package naclutil contains utility functions for working with NaCL key pairs.
package naclutil

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/crypto/nacl/box"
)

const (
	NonceLength = 24
	KeyLength   = 32
)

// Create directory where keys will be stored.
func CreateKeyStore(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "stat failed")
		}
		err = os.MkdirAll(path, 0700)
		if err != nil {
			return errors.Wrap(err, "mkdir failed")
		}
		fi, err = os.Stat(path)
		if err != nil {
			return errors.Wrap(err, "stat2 failed")
		}
	}
	if fi == nil {
		return errors.Wrap(err, "stat returned nil")
	}
	if !fi.IsDir() {
		return errors.Wrap(err, "CreateKeyStore takes a directory")
	}
	return nil
}

// Return keys at the specified location, generating and creating them if necessary.
func FetchKeypair(path, name string) ([]byte, []byte, error) {
	err := CreateKeyStore(path)
	if err != nil {
		return nil, nil, err
	}

	pubPath := fmt.Sprintf("%s/%s.pub", path, name)
	pubInfo, err := os.Stat(pubPath)
	pubExists := true
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, nil, errors.Wrap(err, "stat failed")
		}
		pubExists = false
	}

	keyPath := fmt.Sprintf("%s/%s.key", path, name)
	keyInfo, err := os.Stat(keyPath)
	keyExists := true
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, nil, errors.Wrap(err, "stat failed")
		}
		keyExists = false
	}

	/* Unacceptable: one path exists and other does not */
	if pubExists != keyExists {
		err = errors.Errorf("need neither or both of %s %s\n", pubPath, keyPath)
		return nil, nil, err
	}

	/* Unacceptable: either path is a directory */
	if pubExists && pubInfo.IsDir() {
		err = errors.Errorf("found directory at %s\n", pubPath)
		return nil, nil, err
	}
	if keyExists && keyInfo.IsDir() {
		err = errors.Errorf("found directory at %s\n", keyPath)
		return nil, nil, err
	}

	if !pubExists {
		pub, key, err := box.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, errors.Wrap(err, "key generation failed")
		}

		err = ioutil.WriteFile(pubPath, (*pub)[:], 0644)
		if err != nil {
			return nil, nil, errors.Wrap(err, "public key write failed")
		}

		err = ioutil.WriteFile(keyPath, (*key)[:], 0600)
		if err != nil {
			return nil, nil, errors.Wrap(err, "private key write failed")
		}

		return (*pub)[:], (*key)[:], nil

	} else {
		pub, err := ioutil.ReadFile(pubPath)
		if err != nil {
			return nil, nil, errors.Wrap(err, "public key read failed")
		}
		key, err := ioutil.ReadFile(keyPath)
		if err != nil {
			return nil, nil, errors.Wrap(err, "private key read failed")
		}

		return pub, key, nil
	}
}

func EncryptMessage(msg, _toPub, _fromKey []byte) ([]byte, []byte, error) {
	var nonce [NonceLength]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		return nil, nil, errors.Wrap(err, "nonce read failed")
	}
	/* Copy byte slices into the static arrays that box.Seal expects. */
	var toPub, fromKey [KeyLength]byte
	copy(toPub[:], _toPub)
	copy(fromKey[:], _fromKey)
	enc := box.Seal(nil, msg, &nonce, &toPub, &fromKey)
	return enc, nonce[:], nil
}

func DecryptMessage(enc, _nonce, _fromPub, _toKey []byte) ([]byte, error) {
	var nonce [NonceLength]byte
	copy(nonce[:], _nonce)
	var fromPub, toKey [KeyLength]byte
	copy(fromPub[:], _fromPub)
	copy(toKey[:], _toKey)
	msg, success := box.Open(nil, enc, &nonce, &fromPub, &toKey)
	if !success {
		return nil, errors.Errorf("failed to decrypt message")
	}
	return msg, nil
}
