package infra

import (
	"automation-hub-idp/internal/app/config"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDatabase(user, password, dbName, dbHost string, dbPort int) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		dbHost, user, password, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GetDefaultDB() (*gorm.DB, error) {
	db, err := NewPostgresDatabase(config.PostgresConfig.User, config.PostgresConfig.Password,
		config.PostgresConfig.DbName, config.PostgresConfig.DbHost, config.PostgresConfig.DbPort)
	if err != nil {
		return nil, err
	}

	if err := RunMigrations(db); err != nil {
		return nil, err
	}

	if err := SeedDatabase(db); err != nil {
		return nil, err
	}

	return db, nil
}
