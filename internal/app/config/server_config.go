package config

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	webServerPort string = "WEB_SERVER_PORT"
	baseURL       string = "BASE_URL"
)

type serverConfig struct {
	Port    string
	BaseURL string
}

func newServerConfig() (*serverConfig, error) {
	port := getEnvString(webServerPort, "8080")
	// validate port (if port is just numbers and is between 0 and 65535)
	if len(port) < 1 {
		return nil, errors.New("error: Port is not set, please check the environment variable: " + webServerPort)
	}
	var numPort int
	var err error
	if numPort, err = strconv.Atoi(port); err != nil {
		errorMessage := fmt.Sprintf("error: Port %s is not a valid number port, please check the environment variable: %s", port, webServerPort)
		return nil, errors.New(errorMessage)
	}
	if numPort < 0 || numPort > 65535 {
		errorMessage := fmt.Sprintf("error: Port %d is not valid, please check the environment variable: %s", numPort, webServerPort)
		return nil, errors.New(errorMessage)
	}

	baseURL := getEnvString(baseURL, "/api")

	return &serverConfig{
		Port:    port,
		BaseURL: baseURL,
	}, nil
}
