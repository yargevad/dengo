package main

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
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
	claims := jwtauth.Claims{"user": name}.
		SetExpiryIn(time.Hour * 24).
		SetIssuedNow()
	_, str, err := tokenAuth.Encode(claims)
	return str, err
}

func JWTUser(r *http.Request) string {
	ctx := r.Context()
	token, ok := ctx.Value("jwt").(*jwt.Token)
	if !ok {
		return ""
	}
	claims := token.Claims
	user, ok := claims["user"].(string)
	if !ok {
		return ""
	}
	return user
}
