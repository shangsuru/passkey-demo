package main

import (
	"log"
)

func main() {
	s, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}
	s.Start()
}
