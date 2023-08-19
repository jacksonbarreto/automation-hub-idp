package mappers

import (
	"github.com/stretchr/testify/assert"
	"idp-automations-hub/internal/app/models"
	"testing"
)

func TestUserToUserDTOMapping(t *testing.T) {
	user := models.SimulateUser()

	userResponse, err := MapUserToUserResponse(&user)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, userResponse.ID)
	assert.Equal(t, user.Email, userResponse.Email)
}
