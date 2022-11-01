package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"os"
	"strconv"
)

type Config struct {
	DbUrl            string
	DbPort           int
	DbUserName       string
	DbPassword       string
	DbName           string
	DbMaxConnections int
	BindAddr         string
	BindPort         int
	JwtSecretFile    string
	JwtIcalSecret    string
	JwtPublicKey     []byte
	SslMode          string
	NonceToggle      bool
	XForward         bool
}

func PrintConfHelp() {
	fmt.Println("Create a .env file in the working directory which has the following defined:")
	fmt.Println("DB_URL, DB_PORT, DB_USERNAME, DB_PASSWORD, BIND_ADDR, BIND_PORT, JWT_SECRET_FILE, DB_MAX_CONNS")
	fmt.Println("See ../README.md for more help.")
}

func getEnvVar(EnvVar string) string {
	ret := os.Getenv(EnvVar)
	if ret == "" {
		PrintConfHelp()
		log.Fatalf("Error loading .env file: %s is undefined\n", EnvVar)
	}

	return ret
}

func getEnvVarInt(EnvVar string) int {
	tmp := getEnvVar(EnvVar)
	ret, err := strconv.Atoi(tmp)

	if err != nil {
		PrintConfHelp()
		log.Fatalf("Error loading .env file: %s must be an integer not a string\n", EnvVar)
	}

	return ret
}

func getEnvVarBool(EnvVar string) bool {
	tmp := getEnvVar(EnvVar)
	ret, err := strconv.ParseBool(tmp)
	if err != nil {
		PrintConfHelp()
		log.Fatalf("Error loading .env file: %s must be a boolean.\n", EnvVar)
	}
	return ret
}

func LoadConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		PrintConfHelp()
		log.Println("Error loading .env file, assuming normal env vars.", err)
	}

	ret := Config{DbUrl: getEnvVar("DB_URL"),
		DbPort:           getEnvVarInt("DB_PORT"),
		DbUserName:       getEnvVar("DB_USERNAME"),
		DbName:           getEnvVar("DB_NAME"),
		DbPassword:       getEnvVar("DB_PASSWORD"),
		DbMaxConnections: getEnvVarInt("DB_MAX_CONNS"),
		BindAddr:         getEnvVar("BIND_ADDR"),
		BindPort:         getEnvVarInt("BIND_PORT"),
		JwtIcalSecret:    getEnvVar("JWT_ICAL_SECRET"),
		JwtSecretFile:    getEnvVar("JWT_SECRET_FILE"),
		SslMode:          getEnvVar("SSL_MODE"),
		NonceToggle:      getEnvVarBool("NONCE_TOGGLE"),
		XForward:         getEnvVarBool("XFORWARD")}
	log.Println("Loaded .env file")
	log.Printf("Loading public key from %s\n", ret.JwtSecretFile)

	// Load the jwt secret file into JwtPublicKey
	f, err := os.Open(ret.JwtSecretFile)
	if err != nil {
		log.Println(err)
		return ret, err
	}
	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		log.Println(err)
		return ret, err
	}
	ret.JwtPublicKey = bytes

	return ret, nil
}
