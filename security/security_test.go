package security

import (
	"arcio/attendance-system/config"
	"encoding/hex"
	"github.com/golang-jwt/jwt"
	"io"
	"log"
	"os"
	"testing"
	"time"
)

func TestCheckJwt(t *testing.T) {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Println(err)
		return
	}

	f, err := os.Open("privkey")
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		log.Println(err)
		return
	}

	key, err := jwt.ParseECPrivateKeyFromPEM(bytes)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	// Create a test JWT
	// ED512 is the standard we use.
	claims := Claims{
		Uuid:      "i am a user id, I promise you UwU",
		Type:      "ACCESS",
		Firstname: "Dave",
		Surname:   "Dave",
		Iat:       time.Now().Unix(),
		Exp:       time.Now().Unix() + 9999,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES512, claims)
	signedToken, err := token.SignedString(key)
	signedToken = "Bearer " + signedToken

	if err != nil {
		t.Log("Cannot sign jwt")
		t.Fail()
		return
	}

	_, err = CheckHeaderJwt(signedToken, conf)
	if err != nil {
		t.Log("Cannot verify known good jwt")
		t.Fail()
		return
	}
}

func TestInvalidJwt(t *testing.T) {
	k, _ := hex.DecodeString("aaeaaeaeaeaeaedead")
	conf := config.Config{JwtPublicKey: k}

	signedToken := "Bearer asdfasdf"

	_, err := CheckHeaderJwt(signedToken, conf)
	if err == nil {
		t.Log("Has bearer but is bad")
		t.Fail()
	}

}
