package main

import (
	"context"
	"time"

	"github.com/joho/godotenv"
	"github.com/koltyakov/sp-time-machine/pkg/utils"
	"github.com/koltyakov/sp-time-machine/pkg/worker"

	log "github.com/sirupsen/logrus"
)

var functionTimeout = 600 * 10 // 600 is for Azure Functions

func main() {
	_ = godotenv.Load()

	utils.WithTimeout(time.Duration(functionTimeout-10)*time.Second, func(done context.CancelFunc) {
		if err := worker.Run(); err != nil {
			log.Errorf("Error: %s\n", err)
		}
		done()
	})
}
