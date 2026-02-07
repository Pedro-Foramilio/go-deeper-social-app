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

	addr := env.GetString("DB_ADDR", "postgres://def:def@localhost/socialnetwork?sslmode=disable")
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	store := store.NewStorage(conn)

	db.Seed(store, conn)
}
