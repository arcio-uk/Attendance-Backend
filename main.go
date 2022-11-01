package main

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/middleware"
	"arcio/attendance-system/routes"
	"arcio/attendance-system/security"
	"arcio/attendance-system/utils"
	"fmt"
	"log"
	"runtime"
)

var GlobalConfig config.Config
var DatabasePool *utils.DatabasePool
var NonceManager security.NonceManager

const VERSION_INFO = "Verticle Slice 1"

func main() {
	fmt.Println("Arcio Attendance System Backend, send stdout to a log file for all errors to be logged.")
	fmt.Println(" -> See ./README.md for setup help and the \"arcio-db\" repo for database schemas.")
	fmt.Println(" -> Edit the .env as instructed by ./README.md or tech support to setup.")
	fmt.Printf(" -> Version: \"%s\"\n", VERSION_INFO)
	fmt.Printf(" -> Environment information: \"%s\"\n", runtime.Version())
	fmt.Println("Please send above data in any bug reports or support queries.")
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Println("Starting the attendance server")

	// Read configuration and connect to the database
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Cannot load config")
	}
	GlobalConfig = conf

	database, err := utils.InitDatabasePool(conf)
	if err != nil {
		log.Fatalf("Cannot initialise connection to the database - %s\n", err)
	}
	DatabasePool = database

	// Setup the nonce manager
	log.Println("Starting nonce manager")
	NonceManager.InitNonceManager()

	// Creates a reference to the global nonce manager.
	// See routes/misc.go
	routes.NonceManager = &NonceManager

	// Creates a reference to the global Database Pol.
	// See routes/router.go
	routes.DatabasePool = database

	// Create a reference to the global config
	// See routes/router.go
	routes.GlobalConfig = &GlobalConfig
	middleware.GlobalConfig = &GlobalConfig

	bindAddr := fmt.Sprintf("%s:%d", conf.BindAddr, conf.BindPort)
	log.Printf("Started the attendance server on http://%s\n", bindAddr)

	globalRouter := routes.InitRouter()
	log.Println("Server started")
	globalRouter.Run(bindAddr)
	log.Fatal("Exited for no reason :(")
}
