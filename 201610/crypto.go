package main

import (
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

func JWTString(name string) (string, error) {
	claims := jwtauth.Claims{"uid": name}.
		SetExpiryIn(time.Hour * 24).
		SetIssuedNow()
	_, str, err := tokenAuth.Encode(claims)
	return str, err
}
