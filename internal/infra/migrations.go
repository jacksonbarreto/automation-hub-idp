package infra

import (
	"automation-hub-idp/internal/app/models"
	"automation-hub-idp/internal/app/utils"
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
	if err := db.Where("Email = ?", "admin@example.com").First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			hashedPassword, err := hasher.Hash(defaultPassword)
			if err != nil {
				return err
			}
			adminUser := models.User{
				Email:    defaultEmail,
				Password: hashedPassword,
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
