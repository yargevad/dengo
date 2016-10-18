package main

import (
	"net/http"
	"time"

	"github.com/goware/jwtauth"
	"github.com/yargevad/crypto/naclutil"
)

var pubKey, privKey []byte
var tokenAuth *jwtauth.JwtAuth

func init() {
	var err error
	// generate new or read existing keys
	pubKey, privKey, err = naclutil.FetchKeypair(*keyPath, *keyName)
	if err != nil {
		panic(err)
	}
	// init jwt context using private key
	tokenAuth = jwtauth.New("HS256", privKey, nil)
}

func jwtString(user int) (string, error) {
	claims := jwtauth.Claims{"uid": user}.
		SetExpiryIn(time.Hour * 24).
		SetIssuedNow()
	_, str, err := tokenAuth.Encode(claims)
	return str, err
}

func setLoggedIn(w http.ResponseWriter, user int) {
	/*
		signed, err := jwtString(user)
		if err != nil {
		}
	*/
}
