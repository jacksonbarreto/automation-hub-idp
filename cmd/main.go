package main

import (
	"idp-automations-hub/internal/app/config"
	"idp-automations-hub/internal/app/router"
)

func main() {
	err := config.Setup()
	if err != nil {
		panic(err)
	}

	err = router.Initialize()
	if err != nil {
		panic(err)
	}

}
