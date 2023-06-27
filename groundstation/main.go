package main

import (
	"time"
)

func main() {

	CommunicationManager, err := communication.initializeManager()
	if err != nil {
		panic(err)
	}

	go CommunicationManager.run()

	for {
		time.Sleep(1 * time.Second)
	}

}
