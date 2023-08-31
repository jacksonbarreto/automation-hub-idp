package mappers

import (
	"github.com/mitchellh/mapstructure"
	"idp-automations-hub/internal/app/dto"
	"idp-automations-hub/internal/app/models"
)

func MapUserToUserResponse(user *models.User) (*dto.UserDTO, error) {
	var userResponse dto.UserDTO
	if err := mapstructure.Decode(user, &userResponse); err != nil {
		return nil, err
	}
	return &userResponse, nil
}
