package main

import (
	"log"

	"github.com/Pedro-Foramilio/social/internal/db"
	"github.com/Pedro-Foramilio/social/internal/env"
	"github.com/Pedro-Foramilio/social/internal/store"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load(".env")

	if err != nil {
		log.Printf("Error loading envs: %v\n", err)
	}

	dbConfig := dbConfig{
		addr:         env.GetString("DB_ADDR", "postgres://def:def@localhost/socialnetwork?sslmode=disable"),
		maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
		maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
		maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
	}

	cfg := config{
		addr: env.GetString("ADDR", ":8081"),
		db:   dbConfig,
		env:  env.GetString("ENV", "local"),
	}

	db, err := db.New(
		dbConfig.addr,
		dbConfig.maxOpenConns,
		dbConfig.maxIdleConns,
		dbConfig.maxIdleTime,
	)

	if err != nil {
		log.Panicf("Error connecting to the database: %v\n", err)
	}
	defer db.Close()
	log.Println("Connected to the database successfully")

	store := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
