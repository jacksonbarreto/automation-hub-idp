package service_mock

import (
	"github.com/stretchr/testify/mock"
	"time"
)

type MockBlockListService struct {
	mock.Mock
}

func (m *MockBlockListService) AddToBlockList(jwtUUID string, expirationTime time.Duration) error {
	args := m.Called(jwtUUID, expirationTime)
	return args.Error(0)
}

func (m *MockBlockListService) IsInBlockList(jwtUUID string) (bool, error) {
	args := m.Called(jwtUUID)
	return args.Get(0).(bool), args.Error(1)
}
