package mappers

import (
	"github.com/mitchellh/mapstructure"
	"idp-automations-hub/internal/app/dto"
	"idp-automations-hub/internal/app/models"
)

func MapUserToUserResponse(user *models.User) (*dto.UserResponse, error) {
	var userResponse dto.UserResponse
	if err := mapstructure.Decode(user, &userResponse); err != nil {
		return nil, err
	}
	return &userResponse, nil
}
