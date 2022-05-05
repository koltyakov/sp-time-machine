package main

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// https://dev.to/mattiasfjellstrom/azure-functions-custom-handlers-in-go-logging-31bp

func init() {
	log.SetLevel(log.InfoLevel)
	jsonFormatter := log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: log.FieldMap{
			log.FieldKeyTime:  "timestamp",
			log.FieldKeyMsg:   "message",
			log.FieldKeyLevel: "level",
		},
	}
	log.SetFormatter(&jsonFormatter)
}
