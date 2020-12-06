package main

import (
	"fmt"
	"locallibs/support"
	"time"
)

// Program entry point
func main() {
	support.SetupCloseHandler() // setup ctrl + c to break loop
	fmt.Println("Press ctrl + c to exit...")

	strIP := support.GetOutboundIP().String()
	config := GetConfigContents("support/config.yml")
	clientID := GenerateID()
	//message := TakeUserInput()
	message := "when you compile for user comment this line and uncomment previous line instead"
	timestamp := GetMicroTime()
	keyToWrite := fmt.Sprintf("%s/%s", config.Etcd.BaseKeyToWrite, clientID)
	valueToWrite := fmt.Sprintf("%s | %s | %s", timestamp, strIP, message)

	// if localhost is open use that endpoint instead
	if testSockConnect("127.0.0.1", "2379") {
		config.Etcd.Endpoints = []string{"127.0.0.1:2379"}
		println("Localhost open using localhost instead of config endpoints list")
	}

	fmt.Println("Client ID is: " + clientID)
	WriteToEtcd(config, keyToWrite, valueToWrite)

	readch := make(chan string)
	go ReadEtcdContinuously(readch, config, keyToWrite)

	for true { // loop forever (user expected to break)
		msg := <-readch
		print(msg)

		time.Sleep(3 * time.Second)
	}
}
