package main

import (
	"flag"
	"log"

	"github.com/chmouel/photoschmouel/photos"
)

func main() {
	var static string
	flag.StringVar(&static, "gen", "", "generate all files in static to this dir")
	flag.Parse()
	if static != "" {
		err := photos.MakeStatic(static)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	err := photos.Server()
	if err != nil {
		log.Fatal(err)
	}
}
