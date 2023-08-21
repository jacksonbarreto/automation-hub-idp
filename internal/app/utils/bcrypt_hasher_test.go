package utils

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestBcryptHasher_Hash(t *testing.T) {
	hasher := DefaultBcryptHasher()

	password := "my-secret-password"
	hashedPassword, err := hasher.Hash(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	// Ensure the hashed password is not the same as the plain password
	assert.NotEqual(t, password, hashedPassword)

	// Check the hashed password is valid bcrypt hash
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	assert.NoError(t, err)
}

func TestBcryptHasher_Compare(t *testing.T) {
	hasher := DefaultBcryptHasher()

	password := "my-secret-password"
	hashedPassword, err := hasher.Hash(password)
	assert.NoError(t, err)

	// Comparing the hashed password with the correct original password
	err = hasher.Compare(hashedPassword, password)
	assert.NoError(t, err)

	// Comparing the hashed password with an incorrect password
	err = hasher.Compare(hashedPassword, "wrong-password")
	assert.Error(t, err)
}

func TestDefaultBcryptHasher(t *testing.T) {
	hasher := DefaultBcryptHasher().(*BcryptHasher)

	assert.Equal(t, bcrypt.DefaultCost, hasher.cost)
}
