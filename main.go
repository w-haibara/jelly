package main

import (
	"flag"
	"jelly/handler"
	"log"
	"net/http"
)

func main() {
	path := flag.String("conf", "./conf.json", "path to configration file")
	flag.Parse()
	log.Println("configuration file: ", *path)

	var client *handler.Client
	if err := handler.InitClient(*path, client); err != nil {
		log.Fatal(err)
		return
	}

	http.HandleFunc("/slack/events", client.GetEventsHandler())
	http.HandleFunc("/slack/actions", client.GetActionsHandler())

	log.Println("[INFO] Server listening")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
