package main

import "idp-automations-hub/internal/app/config"

func main() {
	err := config.Setup()
	if err != nil {
		panic(err)
	}
}
