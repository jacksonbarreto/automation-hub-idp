package utils_mock

import "github.com/stretchr/testify/mock"

type MockHasher struct {
	mock.Mock
}

func (m *MockHasher) Compare(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

func (m *MockHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}
