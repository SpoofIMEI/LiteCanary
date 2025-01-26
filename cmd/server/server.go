package main

import (
	metadata "LiteCanary/internal"
	"LiteCanary/internal/database"
	"LiteCanary/internal/server"
	"LiteCanary/internal/server/commands"
	"LiteCanary/internal/server/configuration"
	"log"
	"os"
	"os/signal"

	"github.com/dchest/uniuri"
)

var (
	opts *server.Options
	sc   chan os.Signal // SC (Shutdown Channel)
)

func main() {
	log.Printf(">-LiteCanary Server %s>\n", metadata.Version)

	// Parse config
	var err error
	opts, err = configuration.GetOptions()
	if err != nil {
		log.Fatalf("encountered error while parsing config: %s", err.Error())
	}

	// Init db
	commander, err := commands.New(&database.Opts{
		Location: opts.DatabaseLocation,
		Debug:    opts.Debug,
	})
	opts.Commander = commander
	if err != nil {
		log.Fatalf("encountered error while initializing database: %s", err.Error())
	}

	// Init admin account
	if opts.NoRegistration {
		password := uniuri.NewLen(15)
		err = commander.AddUser("admin", password)
		if err == nil {
			log.Printf("created admin credentials: admin:%s REMEMBER TO CHANGE THE PASSWORD!\n", password)
		}
	}

	// Set up handler for shutdown signal
	sc = make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	go catchInterrupt()

	// Start API
	s := server.New(opts)
	err = s.StartApi()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func catchInterrupt() {
	for range sc {
		log.Println("shutting down!")
		os.Exit(0)
	}
}
