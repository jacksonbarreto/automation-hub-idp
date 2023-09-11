package main

import (
	"automation-hub-idp/internal/app/config"
	"automation-hub-idp/internal/app/router"
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
