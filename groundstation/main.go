package main

import (
	"time"

	"github.com/kukykuk-navigation/communication"
)

func main() {

	CommunicationManager, err := communication.Communication_initializeCommunicationManager()
	if err != nil {
		panic(err)
	}

	go CommunicationManager.run()

	for {
		time.Sleep(1 * time.Second)
	}

}
