package main

import (
	"fmt"
	"os"
)

func main() {
	var clientId = os.Getenv("clientId")
	var apiKey = os.Getenv("apiKey")

	// load yaml file

	fmt.Println(clientId, apiKey)
}
