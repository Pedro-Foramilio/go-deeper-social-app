package main

import (
	"log"

	"github.com/Pedro-Foramilio/social/internal/env"
	"github.com/Pedro-Foramilio/social/internal/store"
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

	store := store.NewStorage(nil)

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
