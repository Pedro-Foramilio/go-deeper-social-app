package main

import (
	"log"

	"github.com/Pedro-Foramilio/social/internal/env"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load(".env")

	if err != nil {
		log.Printf("Error loading envs: %v\n", err)
	}

	cfg := config{
		addr: env.GetString("ADDR", ":8081"),
	}

	app := &application{
		config: cfg,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
