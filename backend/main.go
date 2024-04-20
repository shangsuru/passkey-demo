package main

import (
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	s, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}
	s.Start()
}
