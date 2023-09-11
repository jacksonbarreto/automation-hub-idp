package mappers

import (
	"automation-hub-idp/internal/app/dto"
	"automation-hub-idp/internal/app/models"
	"github.com/mitchellh/mapstructure"
)

func MapUserToUserResponse(user *models.User) (*dto.UserDTO, error) {
	var userResponse dto.UserDTO
	if err := mapstructure.Decode(user, &userResponse); err != nil {
		return nil, err
	}
	return &userResponse, nil
}
