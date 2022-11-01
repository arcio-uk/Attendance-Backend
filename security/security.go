package security

import (
	"arcio/attendance-system/config"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

const AUTH_HEADER = "Authorization"

/*has subject, type and, expires. See the api docs for help.*/
type Claims struct {
	Type      string `json:"type"`
	Uuid      string `json:"uuid"`
	Iat       int64  `json:"iat"`
	Exp       int64  `json:"exp"`
	Firstname string `json:"firstname"`
	Surname   string `json:"surname"`
	jwt.StandardClaims
}

// Time is in milliseconds instead of seconds by default in the auth system.
func (c Claims) Valid() error {
	valid := c.Type != "" && c.Uuid != "" && c.Firstname != "" && c.Surname != "" && c.Exp >= time.Now().Unix() && c.Iat <= time.Now().Unix()
	if !valid {
		return errors.New("One of the claims was invalid - Check JWT claims.")
	}

	return nil
}

func CheckHeaderJwt(JwtIn string, conf config.Config) (Claims, error) {
	splitJwt := strings.Split(JwtIn, "Bearer ")
	if len(splitJwt) != 2 {
		return Claims{}, errors.New("No Bearer token found")
	}
	return CheckJwt(splitJwt[1], conf)
}

func CheckIcalJwt(JwtIn string, conf config.Config) (Claims, error) {
	ret := Claims{}

	_, err := jwt.ParseWithClaims(
		JwtIn,
		&ret,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Bad alg")
			}

			return []byte(conf.JwtIcalSecret), nil
		},
	)

	if err != nil {
		log.Printf("An error %s occurred when checking **REDACTED**\n", err)
		return Claims{}, err
	}

	if ret.Type != "ICAL" {
		log.Println("The jwt is not an ical token, this is forbidden")
		return Claims{}, errors.New("The jwt is not an ical token. This is forbidden")
	}
	return ret, nil
}

func CheckJwt(JwtIn string, conf config.Config) (Claims, error) {
	ret := Claims{}
	key, err := jwt.ParseECPublicKeyFromPEM(conf.JwtPublicKey)

	if err != nil {
		log.Printf("Aerrorn error %s occurred parsing key for JWT\n", err)
		return Claims{}, err
	}

	_, err = jwt.ParseWithClaims(
		JwtIn,
		&ret,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, errors.New("Bad alg")
			}

			return key, nil
		},
	)

	if err != nil {
		log.Printf("An error %s occurred when checking **REDACTED**\n", err)
		return Claims{}, err
	}

	if ret.Type != "ACCESS" {
		log.Println("The jwt is not an access token, this is forbidden")
		return Claims{}, errors.New("The jwt is not an access token. This is forbidden")
	}
	return ret, nil
}
