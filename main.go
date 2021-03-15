package main

import (
	"log"

	"github.com/chmouel/photoschmouel/photos"
)

func main() {
	err := photos.Server()
	if err != nil {
		log.Fatal(err)
	}
}
