package main

import (
	"flag"
	"jelly/handler"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func main() {
	path := flag.String("conf", "./conf.json", "path to configration file")
	flag.Parse()
	log.WithFields(log.Fields{
		"path": *path,
	}).Info("Loading configuration file")

	var client handler.Client
	if err := handler.InitClient(*path, &client); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("InitClient failed")
		return
	}

	http.HandleFunc("/slack/events", client.GetEventsHandler())
	http.HandleFunc("/slack/actions", client.GetActionsHandler())

	log.Info("Server listening")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("ListenAndServe failed")
	}
}
