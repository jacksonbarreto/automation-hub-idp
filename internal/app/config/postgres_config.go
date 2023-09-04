package config

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"idp-automations-hub/internal/infra"
)

const (
	userDb     string = "USER_DB"
	passwordDb string = "PASSWORD_DB"
	dbName     string = "DB_NAME"
	dbHost     string = "DB_HOST"
	dbPort     string = "DB_PORT"
)

type postgresConfig struct {
	User     string
	Password string
	DbName   string
	DbHost   string
	DbPort   int
}

func newPostgresConfig() (*postgresConfig, error) {

	port := getEnvInt(dbPort, -1)
	if port == -1 {
		errorMessage := fmt.Sprintf("error: Port is not a valid number port, please check the environment variable: %s", dbPort)
		return nil, errors.New(errorMessage)
	}
	if port < 0 || port > 65535 {
		errorMessage := fmt.Sprintf("error: Port %d is not valid, please check the environment variable: %s", port, dbPort)
		return nil, errors.New(errorMessage)
	}
	host := getEnvString(dbHost, "NULL")
	if host == "NULL" {
		errorMessage := fmt.Sprintf("error: Host is not set, please check the environment variable: %s", dbHost)
		return nil, errors.New(errorMessage)
	}
	name := getEnvString(dbName, "NULL")
	if name == "NULL" {
		errorMessage := fmt.Sprintf("error: Name is not set, please check the environment variable: %s", dbName)
		return nil, errors.New(errorMessage)
	}

	return &postgresConfig{
		User:     getEnvString(userDb, ""),
		Password: getEnvString(passwordDb, ""),
		DbName:   name,
		DbHost:   host,
		DbPort:   port,
	}, nil
}

func GetDefaultDB() (*gorm.DB, error) {
	db, err := infra.NewPostgresDatabase(PostgresConfig.User, PostgresConfig.Password, PostgresConfig.DbName, PostgresConfig.DbHost, PostgresConfig.DbPort)
	if err != nil {
		return nil, err
	}
	return db, nil
}