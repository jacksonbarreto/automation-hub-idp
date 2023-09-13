package infra

import (
	"automation-hub-idp/internal/app/models"
	"automation-hub-idp/internal/app/utils"
	"errors"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.User{}); err != nil {
		return err
	}
	return nil
}

func SeedDatabase(db *gorm.DB) error {
	hasher := utils.DefaultBcryptHasher()
	defaultPassword := "1234"
	defaultEmail := "admin@admin.nl"
	var user models.User
	err := db.Where("Email = ?", defaultEmail).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			hashedPassword, err := hasher.Hash(defaultPassword)
			if err != nil {
				return err
			}
			adminUser := models.User{
				Email:       defaultEmail,
				Password:    hashedPassword,
				FirstAccess: false,
			}
			if err := db.Create(&adminUser).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
